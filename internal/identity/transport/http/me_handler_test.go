package identityhttp

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iwasawa/hogedd-api/internal/identity"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
)

func TestMeHandlerReturnsAuthenticatedIdentity(t *testing.T) {
	authenticated, err := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	if err != nil {
		t.Fatalf("identity.New() error = %v", err)
	}
	responder := httpapi.NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := NewMeHandler(responder)
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req = req.WithContext(identity.NewContext(req.Context(), authenticated))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Issuer  string `json:"issuer"`
		Subject string `json:"subject"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Issuer != authenticated.Issuer() || body.Subject != authenticated.Subject() {
		t.Errorf("body = %+v", body)
	}
}

func TestMeHandlerMethodContract(t *testing.T) {
	authenticated, err := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	if err != nil {
		t.Fatalf("identity.New() error = %v", err)
	}
	responder := httpapi.NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := NewMeHandler(responder)

	for _, tt := range []struct {
		method     string
		wantStatus int
		wantBody   bool
	}{
		{method: http.MethodHead, wantStatus: http.StatusOK},
		{method: http.MethodPost, wantStatus: http.StatusMethodNotAllowed, wantBody: true},
	} {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/v1/me", nil)
			req = req.WithContext(identity.NewContext(req.Context(), authenticated))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if (rec.Body.Len() > 0) != tt.wantBody {
				t.Errorf("body length = %d, wantBody %v", rec.Body.Len(), tt.wantBody)
			}
		})
	}
}
