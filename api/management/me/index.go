package handler

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/iwasawa/hogedd-api/internal/app"
	"github.com/iwasawa/hogedd-api/internal/config"
	identityauth0 "github.com/iwasawa/hogedd-api/internal/identity/infrastructure/auth0"
	platformpostgres "github.com/iwasawa/hogedd-api/internal/platform/postgres"
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
	userpostgres "github.com/iwasawa/hogedd-api/internal/user/infrastructure/postgres"
)

var application = sync.OnceValue(newApplication)

func newApplication() *app.Application {
	runtimeConfig := config.LoadRuntime()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: runtimeConfig.LogLevel}))
	authenticationConfig, err := config.LoadAuthentication()
	if err != nil {
		panic(err)
	}
	databaseConfig, err := config.LoadDatabase()
	if err != nil {
		panic(err)
	}
	verifier, err := identityauth0.NewVerifier(authenticationConfig)
	if err != nil {
		panic(err)
	}
	database, err := platformpostgres.Open(context.Background(), databaseConfig)
	if err != nil {
		panic(err)
	}
	users := userpostgres.NewUserRegistrar(database)
	application, err := app.New(
		logger,
		app.WithAccessTokenVerifier(verifier),
		app.WithManagementUserGetter(userapp.NewGetManagementUserUseCase(users)),
	)
	if err != nil {
		panic(err)
	}
	return application
}

// Handler は運営User確認Functionの入口です。
func Handler(w http.ResponseWriter, r *http.Request) {
	application().ManagementUserHandler().ServeHTTP(w, r)
}
