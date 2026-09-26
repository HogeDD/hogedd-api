package handler

import (
	"net/http"

	"github.com/iwasawa/hogedd-api/api/apps/shared"
)

// Handler は匿名App起動を記録するVercel Functionです。
func Handler(w http.ResponseWriter, r *http.Request) {
	r.SetPathValue("slug", r.URL.Query().Get("slug"))
	shared.Application().AppLaunchHandler().ServeHTTP(w, r)
}
