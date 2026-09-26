package handler

import (
	"net/http"

	"github.com/iwasawa/hogedd-api/api/apps/shared"
)

var serveAppDetail = func(w http.ResponseWriter, r *http.Request) {
	shared.Application().AppDetailHandler().ServeHTTP(w, r)
}

// Handler はrewriteされたslugをpath valueへ渡す公開アプリ詳細の入口です。
func Handler(w http.ResponseWriter, r *http.Request) {
	r.SetPathValue("slug", r.URL.Query().Get("slug"))
	serveAppDetail(w, r)
}
