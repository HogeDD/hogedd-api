package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerPassesRewrittenSlug(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler(rec, httptest.NewRequest(http.MethodGet, "/api/apps/detail?slug=clean-tasks", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"slug":"clean-tasks"`) {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
}
