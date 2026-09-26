package handler

import (
	"net/http"

	managementapps "github.com/iwasawa/hogedd-api/api/management/apps"
)

// Handler は運営App詳細Functionの入口です。
func Handler(w http.ResponseWriter, r *http.Request) {
	r.SetPathValue("slug", r.URL.Query().Get("slug"))
	managementapps.Application().ManagementAppHandler().ServeHTTP(w, r)
}
