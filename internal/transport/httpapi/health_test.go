package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iwasawa/hogedd-api/internal/health"
)

type healthServiceStub struct{}

func (healthServiceStub) Liveness(context.Context) health.Status {
	return health.Status{Status: "ok"}
}

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := Chain(
		NewHealthHandler(healthServiceStub{}, NewResponder(logger)),
		RequestID(),
		SecurityHeaders(),
	)

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantBody   bool
	}{
		{name: "GET", method: http.MethodGet, wantStatus: http.StatusOK, wantBody: true},
		{name: "HEAD", method: http.MethodHead, wantStatus: http.StatusOK},
		{name: "POST", method: http.MethodPost, wantStatus: http.StatusMethodNotAllowed, wantBody: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "/api/health", nil)
			req.Header.Set("X-Request-ID", "test-request-id")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("X-Request-ID"); got != "test-request-id" {
				t.Errorf("X-Request-ID = %q", got)
			}
			if got := rec.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q", got)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q", got)
			}
			if !tt.wantBody && rec.Body.Len() != 0 {
				t.Errorf("body = %q, want empty", rec.Body.String())
			}
		})
	}
}

func TestHealthResponseContract(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHealthHandler(healthServiceStub{}, NewResponder(logger))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("status body = %q, want %q", body.Status, "ok")
	}
}
