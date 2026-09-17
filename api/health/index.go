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
	return app.New(logger)
}

// Handler はVercel Functionsから呼び出されるヘルスチェックのエントリーポイントです。
func Handler(w http.ResponseWriter, r *http.Request) {
	application.HealthHandler().ServeHTTP(w, r)
}
