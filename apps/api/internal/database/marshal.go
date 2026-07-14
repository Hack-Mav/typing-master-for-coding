package database

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// marshalEntity converts a struct into a JSON payload preserving datastore-tagged
// fields that are hidden from API responses (json:"-"). This lets the database
// store the complete entity while keeping json tags for API responses.
func marshalEntity(src interface{}) ([]byte, error) {
	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return json.Marshal(src)
	}

	out := make(map[string]interface{})
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Type().Field(i)
		if !field.IsExported() {
			continue
		}

		key := entityFieldKey(field)
		if key == "" {
			continue
		}
		out[key] = rv.Field(i).Interface()
	}

	return json.Marshal(out)
}

// unmarshalEntity decodes a JSON payload into dst, trying datastore tag names,
// json tag names, and snake_case field names.
func unmarshalEntity(data []byte, dst interface{}) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return json.Unmarshal(data, dst)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	elem := rv.Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Type().Field(i)
		if !field.IsExported() {
			continue
		}

		for _, key := range entityFieldKeys(field) {
			if payload, ok := raw[key]; ok {
				if err := json.Unmarshal(payload, elem.Field(i).Addr().Interface()); err != nil {
					return fmt.Errorf("database: failed to decode field %s from key %s: %w", field.Name, key, err)
				}
				break
			}
		}
	}

	return nil
}

// entityFieldKey returns the name a field should be stored under. It prefers
// the datastore tag, then the json tag, then the snake_case field name.
func entityFieldKey(field reflect.StructField) string {
	if tag := field.Tag.Get("datastore"); tag != "" {
		name := strings.Split(tag, ",")[0]
		if name != "-" {
			return name
		}
	}

	if tag := field.Tag.Get("json"); tag != "" {
		name := strings.Split(tag, ",")[0]
		if name != "-" {
			return name
		}
	}

	return toSnakeCase(field.Name)
}

// entityFieldKeys returns all possible JSON keys for a field, ordered by
// preference (datastore, json, snake_case field name).
func entityFieldKeys(field reflect.StructField) []string {
	var keys []string
	if tag := field.Tag.Get("datastore"); tag != "" {
		name := strings.Split(tag, ",")[0]
		if name != "-" {
			keys = append(keys, name)
		}
	}
	if tag := field.Tag.Get("json"); tag != "" {
		name := strings.Split(tag, ",")[0]
		if name != "-" {
			keys = append(keys, name)
		}
	}
	keys = append(keys, toSnakeCase(field.Name))
	keys = append(keys, field.Name)
	return keys
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := rune(s[i-1])
				nextLower := i+1 < len(s) && unicode.IsLower(rune(s[i+1]))
				if unicode.IsLower(prev) || (unicode.IsUpper(prev) && nextLower) {
					b.WriteByte('_')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
