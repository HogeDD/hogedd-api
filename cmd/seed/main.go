package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/iwasawa/hogedd-api/internal/config"
	platformpostgres "github.com/iwasawa/hogedd-api/internal/platform/postgres"
	userpostgres "github.com/iwasawa/hogedd-api/internal/user/infrastructure/postgres"
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

	var identity *userpostgres.LocalSeedIdentity
	if cfg.IncludeIdentity {
		identity = &userpostgres.LocalSeedIdentity{
			AuthIssuer:  cfg.AuthIssuer,
			AuthSubject: cfg.AuthSubject,
			Email:       cfg.Email,
			DisplayName: cfg.DisplayName,
		}
	}
	if err := userpostgres.SeedLocalDevelopment(context.Background(), database, identity); err != nil {
		slog.Error("seed local database", "error", err)
		os.Exit(1)
	}
	slog.Info("local database seed completed")
}
