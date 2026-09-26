package handler

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/iwasawa/hogedd-api/internal/app"
	"github.com/iwasawa/hogedd-api/internal/config"
)

var application = newApplication()

func newApplication() *app.Application {
	runtimeConfig := config.LoadRuntime()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: runtimeConfig.LogLevel}))
	application, err := app.New(logger)
	if err != nil {
		panic(err)
	}
	return application
}

// Handler はVercel Functionsから呼び出される公開アプリ一覧の入口です。
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("recommended") == "true" {
		application.AppsRecommendedHandler().ServeHTTP(w, r)
		return
	}
	application.AppsListHandler().ServeHTTP(w, r)
}
