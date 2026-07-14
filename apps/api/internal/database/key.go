package database

import "strconv"

// Key identifies a single entity in the database.
// It is intentionally datastore-compatible so the rest of the codebase can keep
// using NameKey/Kind/Name without importing cloud.google.com/go/datastore.
type Key struct {
	Kind   string
	Name   string
	ID     int64
	Parent *Key
}

// Encode returns the string identifier for the key.
func (k *Key) Encode() string {
	if k == nil {
		return ""
	}
	if k.Name != "" {
		return k.Name
	}
	if k.ID != 0 {
		return strconv.FormatInt(k.ID, 10)
	}
	return ""
}

// NameKey creates a new key that uses the supplied string name.
func NameKey(kind, name string, parent *Key) *Key {
	return &Key{
		Kind:   kind,
		Name:   name,
		Parent: parent,
	}
}

// StringKey is an alias for NameKey.
func StringKey(kind, name string, parent *Key) *Key {
	return NameKey(kind, name, parent)
}

// IncompleteKey creates a new key that will be assigned a name when Put is called.
func IncompleteKey(kind string, parent *Key) *Key {
	return &Key{
		Kind:   kind,
		Parent: parent,
	}
}
