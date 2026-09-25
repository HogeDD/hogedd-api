package handler

import (
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/iwasawa/hogedd-api/internal/app"
	"github.com/iwasawa/hogedd-api/internal/config"
)

var application = sync.OnceValue(newApplication)

func newApplication() *app.Application {
	runtimeConfig := config.LoadRuntime()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: runtimeConfig.LogLevel}))
	application, err := app.New(logger)
	if err != nil {
		panic(err)
	}
	return application
}

// Handler は未知の運営routeへ秘匿404を返します。
func Handler(w http.ResponseWriter, r *http.Request) {
	application().ManagementNotFoundHandler().ServeHTTP(w, r)
}
