package app

import (
	"log/slog"
	"net/http"

	"github.com/iwasawa/hogedd-api/internal/health"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
)

// Application はアプリケーション全体の依存関係を保持するコンポジションルートです。
// 具体的な実装の組み立てをこの型へ集約し、各機能から依存生成の責務を分離します。
type Application struct {
	handler       http.Handler
	healthHandler http.Handler
}

// New はロガーを受け取り、実行可能なApplicationを構築します。
// loggerがnilの場合はslogのデフォルトロガーを使用します。
func New(logger *slog.Logger) *Application {
	if logger == nil {
		logger = slog.Default()
	}

	responder := httpapi.NewResponder(logger)
	healthService := health.NewService()
	healthHandler := httpapi.NewHealthHandler(healthService, responder)
	wrap := func(handler http.Handler) http.Handler {
		return httpapi.Chain(
			handler,
			httpapi.RequestID(),
			httpapi.SecurityHeaders(),
			httpapi.AccessLog(logger),
			httpapi.Recover(logger, responder),
		)
	}

	return &Application{
		handler:       wrap(httpapi.NewRouter(healthHandler)),
		healthHandler: wrap(healthHandler),
	}
}

// Handler はローカルサーバで全ルートを提供するHTTPハンドラーを返します。
func (a *Application) Handler() http.Handler {
	return a.handler
}

// HealthHandler はVercelのヘルスチェックFunctionで使用するHTTPハンドラーを返します。
func (a *Application) HealthHandler() http.Handler {
	return a.healthHandler
}
