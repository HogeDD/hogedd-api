package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iwasawa/hogedd-api/internal/identity"
)

type accessTokenVerifierStub struct {
	wantToken string
	identity  identity.Identity
	err       error
}

func (s accessTokenVerifierStub) Verify(_ context.Context, token string) (identity.Identity, error) {
	if token != s.wantToken {
		return identity.Identity{}, errors.New("unexpected token")
	}
	return s.identity, s.err
}

func TestAuthenticateBearerStoresVerifiedIdentity(t *testing.T) {
	authenticated, err := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	if err != nil {
		t.Fatalf("identity.New() error = %v", err)
	}
	verifier := accessTokenVerifierStub{wantToken: "valid-token", identity: authenticated}
	responder := NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := AuthenticateBearer(verifier, responder)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := identity.FromContext(r.Context())
		if !ok || got != authenticated {
			t.Errorf("context Identity = %+v, %v", got, ok)
		}
		if token, ok := AccessTokenFromContext(r.Context()); !ok || token != "valid-token" {
			t.Errorf("context Access Token = %q, %v", token, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestAuthenticateBearerRejectsInvalidCredentials(t *testing.T) {
	responder := NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	verifier := accessTokenVerifierStub{wantToken: "valid-token", err: errors.New("invalid token")}
	tests := []struct {
		name        string
		authorizers []string
	}{
		{name: "missing"},
		{name: "wrong scheme", authorizers: []string{"Basic credentials"}},
		{name: "missing token", authorizers: []string{"Bearer"}},
		{name: "extra value", authorizers: []string{"Bearer valid-token extra"}},
		{name: "multiple headers", authorizers: []string{"Bearer valid-token", "Bearer valid-token"}},
		{name: "verifier rejection", authorizers: []string{"Bearer valid-token"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := AuthenticateBearer(verifier, responder)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("protected handler was called")
			}))
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			for _, value := range tt.authorizers {
				req.Header.Add("Authorization", value)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
			if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
				t.Errorf("WWW-Authenticate = %q, want Bearer", got)
			}
		})
	}
}

func TestConcealBearerReturnsNotFoundWithoutAuthenticationChallenge(t *testing.T) {
	responder := NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	verifier := accessTokenVerifierStub{wantToken: "valid-token", err: errors.New("invalid token")}
	for _, headers := range [][]string{nil, {"Bearer invalid-token"}, {"Bearer valid-token"}} {
		handler := ConcealBearer(verifier, responder)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("protected handler was called")
		}))
		req := httptest.NewRequest(http.MethodGet, "/v1/management/me", nil)
		for _, value := range headers {
			req.Header.Add("Authorization", value)
		}
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
		if got := rec.Header().Get("WWW-Authenticate"); got != "" {
			t.Errorf("WWW-Authenticate = %q, want empty", got)
		}
	}
}
