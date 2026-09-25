package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/iwasawa/hogedd-api/internal/config"
	platformpostgres "github.com/iwasawa/hogedd-api/internal/platform/postgres"
	userpostgres "github.com/iwasawa/hogedd-api/internal/user/infrastructure/postgres"
	userseedjson "github.com/iwasawa/hogedd-api/internal/user/infrastructure/seedjson"
)

func main() {
	cfg, err := config.LoadLocalSeed()
	if err != nil {
		slog.Error("load local seed configuration", "error", err)
		os.Exit(1)
	}
	database, err := platformpostgres.Open(context.Background(), cfg.Database)
	if err != nil {
		slog.Error("open local seed database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	users, err := userseedjson.Load(cfg.UsersFile)
	if err != nil {
		slog.Error("load local seed users", "error", err)
		os.Exit(1)
	}
	if err := userpostgres.SeedLocalDevelopment(context.Background(), database, users); err != nil {
		slog.Error("seed local database", "error", err)
		os.Exit(1)
	}
	slog.Info("local database seed completed")
}
