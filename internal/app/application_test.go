package app

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iwasawa/hogedd-api/internal/identity"
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
)

type accessTokenVerifierStub struct {
	identity identity.Identity
	err      error
}

type userRegistrarStub struct{}

type userGetterStub struct{}

func (userGetterStub) Execute(context.Context, identity.Identity) (userapp.RegisteredUserResult, error) {
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	return userapp.RegisteredUserResult{
		ID: "0199-user", Email: "owner@example.com", EmailVerified: true,
		Role: "member", Status: "active", CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (userRegistrarStub) Execute(
	context.Context,
	identity.Identity,
	string,
) (userapp.RegisteredUserResult, error) {
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	return userapp.RegisteredUserResult{
		ID: "0199-user", Email: "owner@example.com", EmailVerified: true,
		Role: "member", Status: "active", CreatedAt: now, UpdatedAt: now, Created: true,
	}, nil
}

func (s accessTokenVerifierStub) Verify(context.Context, string) (identity.Identity, error) {
	return s.identity, s.err
}

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
		{path: "/v1/me", wantStatus: http.StatusUnauthorized},
		{path: "/api/me", wantStatus: http.StatusUnauthorized},
		{path: "/v1/users/me", wantStatus: http.StatusUnauthorized},
		{path: "/api/users/me", wantStatus: http.StatusUnauthorized},
		{path: "/v1/users/me/profile", wantStatus: http.StatusUnauthorized},
		{path: "/api/users/me/profile", wantStatus: http.StatusUnauthorized},
		{path: "/v1/management/me", wantStatus: http.StatusNotFound},
		{path: "/api/management/me", wantStatus: http.StatusNotFound},
		{path: "/v1/management/unknown", wantStatus: http.StatusNotFound},
		{path: "/api/management/unknown", wantStatus: http.StatusNotFound},
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

type managementUserGetterStub struct {
	result userapp.ManagementUserResult
	err    error
}

func (s managementUserGetterStub) Execute(context.Context, identity.Identity) (userapp.ManagementUserResult, error) {
	return s.result, s.err
}

func TestApplicationConcealsAndAllowsManagementRoute(t *testing.T) {
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	application, err := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithAccessTokenVerifier(accessTokenVerifierStub{identity: authenticated}),
		WithManagementUserGetter(managementUserGetterStub{result: userapp.ManagementUserResult{ID: "user-1", Role: "owner"}}),
	)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/v1/management/me", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()

	application.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestApplicationProtectsUserRegistrationRoute(t *testing.T) {
	t.Parallel()
	authenticated, err := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	if err != nil {
		t.Fatal(err)
	}
	application, err := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithAccessTokenVerifier(accessTokenVerifierStub{identity: authenticated}),
		WithUserGetter(userGetterStub{}),
		WithUserRegistrar(userRegistrarStub{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPut, "/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()

	application.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestApplicationGetsCurrentUser(t *testing.T) {
	t.Parallel()
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	application, err := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithAccessTokenVerifier(accessTokenVerifierStub{identity: authenticated}),
		WithUserGetter(userGetterStub{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()

	application.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestApplicationProtectsMeRoute(t *testing.T) {
	t.Parallel()

	authenticated, err := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	if err != nil {
		t.Fatalf("identity.New() error = %v", err)
	}
	application, err := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithAccessTokenVerifier(accessTokenVerifierStub{identity: authenticated}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	application.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Issuer  string `json:"issuer"`
		Subject string `json:"subject"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Issuer != authenticated.Issuer() || body.Subject != authenticated.Subject() {
		t.Errorf("body = %+v", body)
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
