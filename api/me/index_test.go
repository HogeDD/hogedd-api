package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerRequiresAccessToken(t *testing.T) {
	t.Setenv("AUTH0_ISSUER_URL", "https://hogedd.jp.auth0.com/")
	t.Setenv("AUTH0_AUDIENCE", "https://api.hogedd.com")
	rec := httptest.NewRecorder()

	Handler(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
