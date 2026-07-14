package database

import "errors"

// ErrNoSuchEntity is returned when an entity does not exist in the database.
var ErrNoSuchEntity = errors.New("database: no such entity")
