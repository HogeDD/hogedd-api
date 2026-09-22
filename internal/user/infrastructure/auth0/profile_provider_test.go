package auth0

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProfileProviderFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/userinfo" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Errorf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"auth0|owner","email":"owner@example.com","email_verified":true}`))
	}))
	defer server.Close()

	provider, err := newProfileProvider(server.URL+"/", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	profile, err := provider.Fetch(context.Background(), "access-token")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if profile.Subject != "auth0|owner" || profile.Email != "owner@example.com" || !profile.EmailVerified {
		t.Errorf("Fetch() = %+v", profile)
	}
}

func TestProfileProviderRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, body: `{}`},
		{name: "invalid JSON", status: http.StatusOK, body: `{`},
		{name: "missing email", status: http.StatusOK, body: `{"sub":"auth0|owner"}`},
		{name: "missing subject", status: http.StatusOK, body: `{"email":"owner@example.com"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			provider, err := newProfileProvider(server.URL+"/", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			if _, err := provider.Fetch(context.Background(), "access-token"); err == nil {
				t.Fatal("Fetch() error = nil")
			}
		})
	}
}
