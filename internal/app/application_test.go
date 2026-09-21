package app

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplicationRoutes(t *testing.T) {
	t.Parallel()

	application, err := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	tests := []struct {
		path       string
		wantStatus int
	}{
		{path: "/health", wantStatus: http.StatusOK},
		{path: "/api/health", wantStatus: http.StatusOK},
		{path: "/api/healthz", wantStatus: http.StatusNotFound},
		{path: "/api/missing", wantStatus: http.StatusNotFound},
		{path: "/v1/apps", wantStatus: http.StatusOK},
		{path: "/v1/apps/clean-tasks", wantStatus: http.StatusOK},
		{path: "/v1/apps/chinchin", wantStatus: http.StatusNotFound},
		{path: "/api/apps", wantStatus: http.StatusOK},
		{path: "/api/apps/detail?slug=clean-tasks", wantStatus: http.StatusOK},
		{path: "/api/apps/detail?slug=chinchin", wantStatus: http.StatusNotFound},
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

func TestPublishedAppHTTPContract(t *testing.T) {
	t.Parallel()

	application, err := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantCode   string
		wantBody   bool
	}{
		{name: "list", method: http.MethodGet, path: "/v1/apps", wantStatus: 200, wantBody: true},
		{name: "detail", method: http.MethodGet, path: "/v1/apps/clean-tasks", wantStatus: 200, wantBody: true},
		{name: "preparing", method: http.MethodGet, path: "/v1/apps/nishida", wantStatus: 404, wantCode: "app_not_found", wantBody: true},
		{name: "unknown", method: http.MethodGet, path: "/v1/apps/unknown", wantStatus: 404, wantCode: "app_not_found", wantBody: true},
		{name: "invalid slug", method: http.MethodGet, path: "/v1/apps/INVALID", wantStatus: 404, wantCode: "app_not_found", wantBody: true},
		{name: "unsupported method", method: http.MethodPost, path: "/v1/apps", wantStatus: 405, wantCode: "method_not_allowed", wantBody: true},
		{name: "head list", method: http.MethodHead, path: "/v1/apps", wantStatus: 200},
		{name: "head missing", method: http.MethodHead, path: "/v1/apps/nishida", wantStatus: 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := httptest.NewRecorder()
			application.Handler().ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if !tt.wantBody && rec.Body.Len() != 0 {
				t.Fatalf("HEAD body = %q, want empty", rec.Body.String())
			}
			if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
				t.Fatalf("Content-Type = %q", rec.Header().Get("Content-Type"))
			}
			if tt.wantStatus == 405 && rec.Header().Get("Allow") != "GET, HEAD" {
				t.Fatalf("Allow = %q", rec.Header().Get("Allow"))
			}
			if tt.wantCode != "" {
				var body struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error.Code != tt.wantCode {
					t.Fatalf("error body = %s, decode error = %v", rec.Body.String(), err)
				}
			}
		})
	}

	t.Run("list JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		application.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/apps", nil))
		var body struct {
			Data []struct {
				Slug        string `json:"slug"`
				Status      string `json:"status"`
				PublishedAt string `json:"published_at"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode JSON: %v", err)
		}
		if len(body.Data) != 1 || body.Data[0].Slug != "clean-tasks" || body.Data[0].Status != "published" || body.Data[0].PublishedAt != "2026-05-30T00:00:00Z" {
			t.Fatalf("data = %+v", body.Data)
		}
	})
}
