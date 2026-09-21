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

// Handler はrewriteされたslugをpath valueへ渡す公開アプリ詳細の入口です。
func Handler(w http.ResponseWriter, r *http.Request) {
	r.SetPathValue("slug", r.URL.Query().Get("slug"))
	application.AppDetailHandler().ServeHTTP(w, r)
}
