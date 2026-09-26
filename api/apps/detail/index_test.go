package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerPassesRewrittenSlug(t *testing.T) {
	rec := httptest.NewRecorder()
	original := serveAppDetail
	t.Cleanup(func() { serveAppDetail = original })
	serveAppDetail = func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("slug") != "clean-tasks" {
			t.Fatalf("slug = %q", r.PathValue("slug"))
		}
		w.WriteHeader(http.StatusOK)
	}
	Handler(rec, httptest.NewRequest(http.MethodGet, "/api/apps/detail?slug=clean-tasks", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
