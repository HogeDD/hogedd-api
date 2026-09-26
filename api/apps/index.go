package handler

import (
	"net/http"

	"github.com/iwasawa/hogedd-api/api/apps/shared"
)

// Handler はVercel Functionsから呼び出される公開アプリ一覧の入口です。
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("recommended") == "true" {
		shared.Application().AppsRecommendedHandler().ServeHTTP(w, r)
		return
	}
	shared.Application().AppsListHandler().ServeHTTP(w, r)
}
