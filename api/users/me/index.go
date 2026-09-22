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
	userauth0 "github.com/iwasawa/hogedd-api/internal/user/infrastructure/auth0"
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
	accessTokenVerifier, err := identityauth0.NewVerifier(authenticationConfig)
	if err != nil {
		panic(err)
	}
	profileProvider, err := userauth0.NewProfileProvider(authenticationConfig.IssuerURL)
	if err != nil {
		panic(err)
	}
	database, err := platformpostgres.Open(context.Background(), databaseConfig)
	if err != nil {
		panic(err)
	}
	registerUser := userapp.NewRegisterAuthenticatedUserUseCase(
		profileProvider,
		userpostgres.NewUserRegistrar(database),
	)
	getUser := userapp.NewGetCurrentUserUseCase(userpostgres.NewUserRegistrar(database))
	application, err := app.New(
		logger,
		app.WithAccessTokenVerifier(accessTokenVerifier),
		app.WithUserGetter(getUser),
		app.WithUserRegistrar(registerUser),
	)
	if err != nil {
		panic(err)
	}
	return application
}

// Handler はVercel Functionsから呼び出される認証済みUser登録の入口です。
func Handler(w http.ResponseWriter, r *http.Request) {
	application().UserRegistrationHandler().ServeHTTP(w, r)
}
