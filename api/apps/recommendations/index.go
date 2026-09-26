package handler

import (
	"net/http"

	"github.com/iwasawa/hogedd-api/api/apps/shared"
)

// Handler はApp推薦枠を返すVercel Functionです。
func Handler(w http.ResponseWriter, r *http.Request) {
	shared.Application().AppRecommendationsHandler().ServeHTTP(w, r)
}
