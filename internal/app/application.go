package app

import (
	"log/slog"
	"net/http"

	"github.com/iwasawa/hogedd-api/internal/content/application"
	"github.com/iwasawa/hogedd-api/internal/content/infrastructure"
	contenthttp "github.com/iwasawa/hogedd-api/internal/content/transport/http"
	"github.com/iwasawa/hogedd-api/internal/health"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
)

// Application はアプリケーション全体の依存関係を保持するコンポジションルートです。
// 具体的な実装の組み立てをこの型へ集約し、各機能から依存生成の責務を分離します。
type Application struct {
	handler          http.Handler
	healthHandler    http.Handler
	appsListHandler  http.Handler
	appDetailHandler http.Handler
}

// New はロガーを受け取り、実行可能なApplicationを構築します。
// loggerがnilの場合はslogのデフォルトロガーを使用します。
func New(logger *slog.Logger) (*Application, error) {
	if logger == nil {
		logger = slog.Default()
	}

	responder := httpapi.NewResponder(logger)
	healthService := health.NewService()
	healthHandler := httpapi.NewHealthHandler(healthService, responder)
	appStore, err := infrastructure.NewSeededMemoryAppStore()
	if err != nil {
		return nil, err
	}
	appsHandler := contenthttp.NewAppsHandler(
		application.NewListPublishedAppsUseCase(appStore),
		application.NewGetPublishedAppUseCase(appStore),
		responder,
	)
	middleware := httpapi.NewMiddlewareStack(
		httpapi.RequestID(),
		httpapi.SecurityHeaders(),
		httpapi.AccessLog(logger),
		httpapi.Recover(logger, responder),
	)

	return &Application{
		handler: middleware.Wrap(httpapi.NewRouter(
			healthHandler,
			http.HandlerFunc(appsHandler.List),
			http.HandlerFunc(appsHandler.Get),
		)),
		healthHandler:    middleware.Wrap(healthHandler),
		appsListHandler:  middleware.Wrap(http.HandlerFunc(appsHandler.List)),
		appDetailHandler: middleware.Wrap(http.HandlerFunc(appsHandler.Get)),
	}, nil
}

// Handler はローカルサーバで全ルートを提供するHTTPハンドラーを返します。
func (a *Application) Handler() http.Handler {
	return a.handler
}

// HealthHandler はVercelのヘルスチェックFunctionで使用するHTTPハンドラーを返します。
func (a *Application) HealthHandler() http.Handler {
	return a.healthHandler
}

// AppsListHandler はVercelの公開アプリ一覧Functionで使用するHandlerを返します。
func (a *Application) AppsListHandler() http.Handler {
	return a.appsListHandler
}

// AppDetailHandler はVercelの公開アプリ詳細Functionで使用するHandlerを返します。
func (a *Application) AppDetailHandler() http.Handler {
	return a.appDetailHandler
}
