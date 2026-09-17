package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplicationRoutes(t *testing.T) {
	t.Parallel()

	application := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	tests := []struct {
		path       string
		wantStatus int
	}{
		{path: "/health", wantStatus: http.StatusOK},
		{path: "/api/health", wantStatus: http.StatusOK},
		{path: "/api/healthz", wantStatus: http.StatusNotFound},
		{path: "/api/missing", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			application.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if rec.Header().Get("X-Request-ID") == "" {
				t.Error("X-Request-ID is empty")
			}
		})
	}
}
