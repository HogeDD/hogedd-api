package httpapi

import "net/http"

// NewRouter はローカル実行で使用するAPIルーターを構築します。
// Vercelでは各Functionが対応するハンドラーを直接使用します。
func NewRouter(healthHandler, appsListHandler, appDetailHandler, meHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/api/health", healthHandler)
	mux.Handle("/v1/apps", appsListHandler)
	mux.Handle("/v1/apps/{slug}", appDetailHandler)
	mux.Handle("/v1/me", meHandler)
	mux.Handle("/api/apps", appsListHandler)
	mux.HandleFunc("/api/apps/detail", func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("slug", r.URL.Query().Get("slug"))
		appDetailHandler.ServeHTTP(w, r)
	})
	mux.Handle("/api/me", meHandler)
	return mux
}
