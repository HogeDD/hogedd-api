package httpapi

import "net/http"

// NewRouter はローカル実行で使用するAPIルーターを構築します。
// Vercelでは各Functionが対応するハンドラーを直接使用します。
func NewRouter(healthHandler, appsListHandler, appsRecommendedHandler, appDetailHandler, meHandler, userRegistrationHandler, userProfileHandler, managementUserHandler, managementAppsHandler, managementAppHandler, managementAppPublicationHandler, managementNotFoundHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/api/health", healthHandler)
	mux.Handle("/v1/apps", appsListHandler)
	mux.Handle("/v1/apps/recommended", appsRecommendedHandler)
	mux.Handle("/v1/apps/{slug}", appDetailHandler)
	mux.Handle("/v1/me", meHandler)
	mux.Handle("/v1/users/me", userRegistrationHandler)
	mux.Handle("/v1/users/me/profile", userProfileHandler)
	mux.Handle("/v1/management/me", managementUserHandler)
	mux.Handle("/v1/management/apps", managementAppsHandler)
	mux.Handle("/v1/management/apps/{slug}", managementAppHandler)
	mux.Handle("/v1/management/apps/{slug}/publication", managementAppPublicationHandler)
	mux.Handle("/v1/management", managementNotFoundHandler)
	mux.Handle("/v1/management/", managementNotFoundHandler)
	mux.Handle("/api/apps", appsListHandler)
	mux.Handle("/api/apps/recommended", appsRecommendedHandler)
	mux.HandleFunc("/api/apps/detail", func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("slug", r.URL.Query().Get("slug"))
		appDetailHandler.ServeHTTP(w, r)
	})
	mux.Handle("/api/me", meHandler)
	mux.Handle("/api/users/me", userRegistrationHandler)
	mux.Handle("/api/users/me/profile", userProfileHandler)
	mux.Handle("/api/management/me", managementUserHandler)
	mux.Handle("/api/management/apps", managementAppsHandler)
	mux.HandleFunc("/api/management/apps/detail", func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("slug", r.URL.Query().Get("slug"))
		managementAppHandler.ServeHTTP(w, r)
	})
	mux.HandleFunc("/api/management/apps/publication", func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("slug", r.URL.Query().Get("slug"))
		managementAppPublicationHandler.ServeHTTP(w, r)
	})
	mux.Handle("/api/management", managementNotFoundHandler)
	mux.Handle("/api/management/", managementNotFoundHandler)
	return mux
}
