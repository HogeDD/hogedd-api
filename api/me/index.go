package handler

import (
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/iwasawa/hogedd-api/internal/app"
	"github.com/iwasawa/hogedd-api/internal/config"
	"github.com/iwasawa/hogedd-api/internal/identity/infrastructure/auth0"
)

var application = sync.OnceValue(newApplication)

func newApplication() *app.Application {
	runtimeConfig := config.LoadRuntime()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: runtimeConfig.LogLevel}))
	authenticationConfig, err := config.LoadAuthentication()
	if err != nil {
		panic(err)
	}
	accessTokenVerifier, err := auth0.NewVerifier(authenticationConfig)
	if err != nil {
		panic(err)
	}
	application, err := app.New(logger, app.WithAccessTokenVerifier(accessTokenVerifier))
	if err != nil {
		panic(err)
	}
	return application
}

// Handler はVercel Functionsから呼び出される認証主体取得の入口です。
func Handler(w http.ResponseWriter, r *http.Request) {
	application().MeHandler().ServeHTTP(w, r)
}
