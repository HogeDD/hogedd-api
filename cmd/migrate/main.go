package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/iwasawa/hogedd-api/internal/config"
	"github.com/iwasawa/hogedd-api/internal/platform/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	url, err := config.LoadMigrationDatabaseURL()
	if err != nil {
		slog.Error("load migration configuration", "error", err)
		os.Exit(1)
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		slog.Error("open migration database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := postgres.MigrateUp(context.Background(), db); err != nil {
		slog.Error("apply database migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations completed")
}
