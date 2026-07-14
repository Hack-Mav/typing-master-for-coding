package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"google.golang.org/api/iterator"
)

type postgresStorage struct {
	db *sqlx.DB
}

func newPostgresStorage(db *sqlx.DB) *postgresStorage {
	return &postgresStorage{db: db}
}

func (p *postgresStorage) Put(ctx context.Context, key *Key, src interface{}) (*Key, error) {
	if key == nil {
		return nil, errors.New("database: key is nil")
	}
	if src == nil {
		return nil, errors.New("database: src is nil")
	}

	data, err := marshalEntity(src)
	if err != nil {
		return nil, fmt.Errorf("database: failed to marshal entity: %w", err)
	}

	now := time.Now().UTC()
	const q = `
		INSERT INTO entities (kind, name, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (kind, name)
		DO UPDATE SET data = EXCLUDED.data, updated_at = EXCLUDED.updated_at
		RETURNING name
	`
	var returnedName string
	if err := p.db.QueryRowContext(ctx, q, key.Kind, key.Name, data, now, now).Scan(&returnedName); err != nil {
		return nil, fmt.Errorf("database: failed to put entity: %w", err)
	}
	return key, nil
}

func (p *postgresStorage) PutMulti(ctx context.Context, keys []*Key, src interface{}) ([]*Key, error) {
	v := reflect.ValueOf(src)
	if v.Kind() != reflect.Slice {
		return nil, errors.New("database: src must be a slice")
	}
	if len(keys) != v.Len() {
		return nil, errors.New("database: keys and src length mismatch")
	}

	returnedKeys := make([]*Key, len(keys))
	for i, key := range keys {
		elem := v.Index(i).Interface()
		newKey, err := p.Put(ctx, key, elem)
		if err != nil {
			return nil, err
		}
		returnedKeys[i] = newKey
	}
	return returnedKeys, nil
}

func (p *postgresStorage) Get(ctx context.Context, key *Key, dst interface{}) error {
	if key == nil {
		return ErrNoSuchEntity
	}

	var data json.RawMessage
	const q = `SELECT data FROM entities WHERE kind = $1 AND name = $2`
	if err := p.db.QueryRowxContext(ctx, q, key.Kind, key.Name).Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoSuchEntity
		}
		return fmt.Errorf("database: failed to get entity: %w", err)
	}

	if err := unmarshalEntity(data, dst); err != nil {
		return fmt.Errorf("database: failed to unmarshal entity: %w", err)
	}
	return nil
}

func (p *postgresStorage) GetAll(ctx context.Context, query *Query, dst interface{}) ([]*Key, error) {
	if query.err != nil {
		return nil, query.err
	}

	keys, data, err := p.query(ctx, query)
	if err != nil {
		return nil, err
	}

	if query.keysOnly || dst == nil {
		return keys, nil
	}

	if err := unmarshalSlice(dst, data); err != nil {
		return nil, err
	}
	return keys, nil
}

func (p *postgresStorage) Run(ctx context.Context, query *Query) Iterator {
	keys, data, err := p.query(ctx, query)
	if err != nil {
		return &errorIterator{err: err}
	}
	return newSliceIterator(keys, data)
}

func (p *postgresStorage) Delete(ctx context.Context, key *Key) error {
	if key == nil {
		return nil
	}
	const q = `DELETE FROM entities WHERE kind = $1 AND name = $2`
	_, err := p.db.ExecContext(ctx, q, key.Kind, key.Name)
	return err
}

func (p *postgresStorage) DeleteMulti(ctx context.Context, keys []*Key) error {
	for _, key := range keys {
		if err := p.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (p *postgresStorage) Count(ctx context.Context, query *Query) (int, error) {
	if query.err != nil {
		return 0, query.err
	}

	where, args, err := buildWhere(query)
	if err != nil {
		return 0, err
	}

	count := 0
	q := "SELECT COUNT(*) FROM entities WHERE kind = $1" + where
	args = append([]interface{}{query.kind}, args...)
	if err := p.db.QueryRowxContext(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (p *postgresStorage) query(ctx context.Context, query *Query) ([]*Key, [][]byte, error) {
	where, args, err := buildWhere(query)
	if err != nil {
		return nil, nil, err
	}

	q := "SELECT name, data FROM entities WHERE kind = $1" + where
	args = append([]interface{}{query.kind}, args...)

	for _, ord := range query.orders {
		q += fmt.Sprintf(" ORDER BY data->'%s' %s", toSnakeCase(ord.fieldName), ord.direction())
	}
	if query.limit > 0 {
		q += " LIMIT " + strconv.Itoa(query.limit)
	}
	if query.offset > 0 {
		q += " OFFSET " + strconv.Itoa(query.offset)
	}

	rows, err := p.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var keys []*Key
	var data [][]byte
	for rows.Next() {
		var name string
		var payload json.RawMessage
		if err := rows.Scan(&name, &payload); err != nil {
			return nil, nil, err
		}
		keys = append(keys, &Key{Kind: query.kind, Name: name})
		data = append(data, payload)
	}
	return keys, data, rows.Err()
}

func buildWhere(query *Query) (string, []interface{}, error) {
	var where string
	var args []interface{}
	argIdx := 1

	for _, f := range query.filters {
		field := toSnakeCase(f.fieldName)
		op, ok := normalizeOp(f.operator)
		if !ok {
			return "", nil, fmt.Errorf("database: unsupported operator %q", f.operator)
		}

		var expr string
		if op == "IN" {
			vals, err := sliceValue(f.value)
			if err != nil {
				return "", nil, err
			}
			placeholders := make([]string, len(vals))
			for i, v := range vals {
				args = append(args, v)
				placeholders[i] = "$" + strconv.Itoa(argIdx+1)
				argIdx++
			}
			expr = fmt.Sprintf("data->'%s' IN (%s)", field, strings.Join(placeholders, ", "))
		} else {
			args = append(args, f.value)
			argIdx++
			expr = fmt.Sprintf("data->'%s' %s to_jsonb($%d)", field, op, argIdx)
		}
		where += " AND " + expr
	}

	return where, args, nil
}

func normalizeOp(op string) (string, bool) {
	switch strings.TrimSpace(strings.ToLower(op)) {
	case "=", "==":
		return "=", true
	case ">":
		return ">", true
	case "<":
		return "<", true
	case ">=":
		return ">=", true
	case "<=":
		return "<=", true
	case "!=", "<>":
		return "!=", true
	case "in":
		return "IN", true
	case "not-in", "notin":
		return "NOT IN", true
	}
	return "", false
}

func sliceValue(v interface{}) ([]interface{}, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("database: IN operator requires slice/array value")
	}
	var out []interface{}
	for i := 0; i < rv.Len(); i++ {
		out = append(out, rv.Index(i).Interface())
	}
	return out, nil
}

func (o order) direction() string {
	if o.descending {
		return "DESC"
	}
	return "ASC"
}

func unmarshalSlice(dst interface{}, data [][]byte) error {
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("database: dst must be pointer to slice")
	}
	sliceVal := dv.Elem()
	elemType := sliceVal.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr
	t := elemType
	if isPtr {
		t = t.Elem()
	}

	for _, payload := range data {
		newElem := reflect.New(t)
		if err := unmarshalEntity(payload, newElem.Interface()); err != nil {
			return err
		}
		if isPtr {
			sliceVal = reflect.Append(sliceVal, newElem)
		} else {
			sliceVal = reflect.Append(sliceVal, reflect.Indirect(newElem))
		}
	}
	dv.Elem().Set(sliceVal)
	return nil
}

type errorIterator struct {
	err error
}

func (it *errorIterator) Next(dst interface{}) (*Key, error) {
	return nil, it.err
}

// Ensure errorIterator returns iterator.Done when err is nil to keep the API consistent.
func init() {
	_ = iterator.Done
}
