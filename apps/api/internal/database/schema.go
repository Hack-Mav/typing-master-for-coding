package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// migrateSchema creates the single entity table used to store all kinds
// in a Datastore-compatible way using JSONB.
func migrateSchema(db *sqlx.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS entities (
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (kind, name)
);

CREATE INDEX IF NOT EXISTS idx_entities_kind_data ON entities USING GIN (kind, data);
`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create entities table: %w", err)
	}
	return nil
}
