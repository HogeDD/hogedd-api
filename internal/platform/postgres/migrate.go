package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// NewMigrationProvider は埋め込まれたmigrationを実行するProviderを返します。
func NewMigrationProvider(db *sql.DB) (*goose.Provider, error) {
	migrations, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("load migrations: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations)
	if err != nil {
		return nil, fmt.Errorf("create migration provider: %w", err)
	}
	return provider, nil
}

// MigrateUp は未適用のschema migrationをすべて適用します。
func MigrateUp(ctx context.Context, db *sql.DB) error {
	provider, err := NewMigrationProvider(db)
	if err != nil {
		return err
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("migrate postgres: %w", err)
	}
	if err := verifySchema(ctx, db); err != nil {
		return err
	}
	return nil
}

func verifySchema(ctx context.Context, db *sql.DB) error {
	var usersExists bool
	if err := db.QueryRowContext(ctx, "SELECT to_regclass('public.users') IS NOT NULL").Scan(&usersExists); err != nil {
		return fmt.Errorf("verify postgres schema: %w", err)
	}
	if !usersExists {
		return fmt.Errorf("verify postgres schema: users table is missing")
	}
	var profilesExists bool
	if err := db.QueryRowContext(ctx, "SELECT to_regclass('public.user_profiles') IS NOT NULL").Scan(&profilesExists); err != nil {
		return fmt.Errorf("verify postgres schema: %w", err)
	}
	if !profilesExists {
		return fmt.Errorf("verify postgres schema: user_profiles table is missing")
	}
	var contentAppsExists bool
	if err := db.QueryRowContext(ctx, "SELECT to_regclass('public.content_apps') IS NOT NULL").Scan(&contentAppsExists); err != nil {
		return fmt.Errorf("verify postgres schema: %w", err)
	}
	if !contentAppsExists {
		return fmt.Errorf("verify postgres schema: content_apps table is missing")
	}
	return nil
}
