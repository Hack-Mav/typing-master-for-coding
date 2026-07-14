package database

import (
	"fmt"
	"strings"
)

func trimSpaces(s string) string { return strings.TrimSpace(s) }

func startsWith(s, prefix string) bool { return strings.HasPrefix(s, prefix) }

func trimRightOperators(s string) string {
	return strings.TrimRight(s, " >=<=!")
}

func errInvalidFilter(f string) error {
	return fmt.Errorf("database: invalid query filter %q", f)
}

// Query represents a database query. It is a deliberately small,
// datastore-compatible subset of the original Datastore query builder.
type Query struct {
	kind       string
	ancestor   *Key
	filters    []filter
	orders     []order
	limit      int
	offset     int
	keysOnly   bool
	projection []string
	err        error
}

type filter struct {
	fieldName string
	operator  string
	value     interface{}
}

type order struct {
	fieldName  string
	descending bool
}

// NewQuery creates a new query for a specific entity kind.
func NewQuery(kind string) *Query {
	return &Query{
		kind:  kind,
		limit: -1,
	}
}

// clone returns a deep-ish copy of the query.
func (q *Query) clone() *Query {
	if q == nil {
		return nil
	}
	x := *q
	if len(q.filters) > 0 {
		x.filters = make([]filter, len(q.filters))
		copy(x.filters, q.filters)
	}
	if len(q.orders) > 0 {
		x.orders = make([]order, len(q.orders))
		copy(x.orders, q.orders)
	}
	return &x
}

// Filter appends a field-based filter. The filterStr is a field name followed
// by an optional operator (e.g. "user_id =", "CreatedAt >").
func (q *Query) Filter(filterStr string, value interface{}) *Query {
	q = q.clone()
	filterStr = trimSpaces(filterStr)
	if filterStr == "" {
		q.err = errInvalidFilter(filterStr)
		return q
	}

	fieldName := trimRightOperators(filterStr)
	operator := trimSpaces(filterStr[len(fieldName):])
	if operator == "" {
		operator = "="
	}
	return q.FilterField(fieldName, operator, value)
}

// FilterField appends a field-based filter with explicit operator.
func (q *Query) FilterField(fieldName, operator string, value interface{}) *Query {
	q = q.clone()
	q.filters = append(q.filters, filter{
		fieldName: fieldName,
		operator:  operator,
		value:     value,
	})
	return q
}

// Order adds a sort order. Prefix fieldName with "-" for descending.
func (q *Query) Order(fieldName string) *Query {
	q = q.clone()
	fieldName = trimSpaces(fieldName)
	descending := false
	if startsWith(fieldName, "-") {
		descending = true
		fieldName = trimSpaces(fieldName[1:])
	}
	q.orders = append(q.orders, order{
		fieldName:  fieldName,
		descending: descending,
	})
	return q
}

// Limit caps the number of returned entities.
func (q *Query) Limit(n int) *Query {
	q = q.clone()
	q.limit = n
	return q
}

// Offset skips the first n results.
func (q *Query) Offset(n int) *Query {
	q = q.clone()
	q.offset = n
	return q
}

// KeysOnly configures the query to return only keys.
func (q *Query) KeysOnly() *Query {
	q = q.clone()
	q.keysOnly = true
	return q
}

// Project limits the returned fields to the supplied projection.
func (q *Query) Project(fields ...string) *Query {
	q = q.clone()
	q.projection = append(q.projection, fields...)
	return q
}

// Ancestor restricts the query to entities with the supplied ancestor.
func (q *Query) Ancestor(ancestor *Key) *Query {
	q = q.clone()
	q.ancestor = ancestor
	return q
}

// Error returns any query construction error.
func (q *Query) Error() error {
	return q.err
}
