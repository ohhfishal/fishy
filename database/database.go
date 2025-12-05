package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

func Connect(ctx context.Context, driver string, connection string) (*Queries, error) {
	db, err := sql.Open(driver, connection)
	if err != nil {
		return nil, fmt.Errorf("opening connection: %w", err)
	}

	if err := RunMigrations(ctx, db); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return New(db), nil
}

func RunMigrations(ctx context.Context, db DBTX) error {
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return err
	}
	return nil
}
