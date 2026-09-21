package httpapi

import "net/http"

// NewRouter はローカル実行で使用するAPIルーターを構築します。
// Vercelでは各Functionが対応するハンドラーを直接使用します。
func NewRouter(healthHandler, appsListHandler, appDetailHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/api/health", healthHandler)
	mux.Handle("/v1/apps", appsListHandler)
	mux.Handle("/v1/apps/{slug}", appDetailHandler)
	return mux
}
