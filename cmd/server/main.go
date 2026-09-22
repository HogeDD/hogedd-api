package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/iwasawa/hogedd-api/internal/app"
	"github.com/iwasawa/hogedd-api/internal/config"
	"github.com/iwasawa/hogedd-api/internal/identity/infrastructure/auth0"
	"github.com/iwasawa/hogedd-api/internal/platform/httpserver"
	platformpostgres "github.com/iwasawa/hogedd-api/internal/platform/postgres"
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
	userauth0 "github.com/iwasawa/hogedd-api/internal/user/infrastructure/auth0"
	userpostgres "github.com/iwasawa/hogedd-api/internal/user/infrastructure/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)
	authenticationConfig, authenticationEnabled, err := config.LoadOptionalAuthentication()
	if err != nil {
		logger.Error("load authentication configuration", "error", err)
		os.Exit(1)
	}
	var applicationOptions []app.Option
	if authenticationEnabled {
		accessTokenVerifier, err := auth0.NewVerifier(authenticationConfig)
		if err != nil {
			logger.Error("build access token verifier", "error", err)
			os.Exit(1)
		}
		applicationOptions = append(applicationOptions, app.WithAccessTokenVerifier(accessTokenVerifier))

		databaseConfig, err := config.LoadDatabase()
		if err != nil {
			logger.Error("load database configuration", "error", err)
			os.Exit(1)
		}
		database, err := platformpostgres.Open(context.Background(), databaseConfig)
		if err != nil {
			logger.Error("open database", "error", err)
			os.Exit(1)
		}
		defer database.Close()

		profileProvider, err := userauth0.NewProfileProvider(authenticationConfig.IssuerURL)
		if err != nil {
			logger.Error("build Auth0 profile provider", "error", err)
			os.Exit(1)
		}
		registerUser := userapp.NewRegisterAuthenticatedUserUseCase(
			profileProvider,
			userpostgres.NewUserRegistrar(database),
		)
		applicationOptions = append(applicationOptions, app.WithUserRegistrar(registerUser))
	}

	application, err := app.New(logger, applicationOptions...)
	if err != nil {
		logger.Error("build application", "error", err)
		os.Exit(1)
	}
	server := httpserver.New(cfg, application.Handler(), logger)

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(shutdownSignal); err != nil {
		logger.Error("HTTP server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
