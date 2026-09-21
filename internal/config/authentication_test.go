package config

import "testing"

func TestLoadAuthentication(t *testing.T) {
	t.Setenv("AUTH0_ISSUER_URL", "https://hogedd.jp.auth0.com/")
	t.Setenv("AUTH0_AUDIENCE", "https://api.hogedd.com")

	got, err := LoadAuthentication()
	if err != nil {
		t.Fatalf("LoadAuthentication() error = %v", err)
	}
	if got.IssuerURL != "https://hogedd.jp.auth0.com/" {
		t.Errorf("IssuerURL = %q", got.IssuerURL)
	}
	if got.Audience != "https://api.hogedd.com" {
		t.Errorf("Audience = %q", got.Audience)
	}
}

func TestLoadAuthenticationRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		issuer   string
		audience string
	}{
		{name: "missing issuer", audience: "https://api.hogedd.com"},
		{name: "non HTTPS issuer", issuer: "http://hogedd.jp.auth0.com/", audience: "https://api.hogedd.com"},
		{name: "relative issuer", issuer: "hogedd.jp.auth0.com/", audience: "https://api.hogedd.com"},
		{name: "issuer without trailing slash", issuer: "https://hogedd.jp.auth0.com", audience: "https://api.hogedd.com"},
		{name: "issuer with query", issuer: "https://hogedd.jp.auth0.com/?tenant=hogedd", audience: "https://api.hogedd.com"},
		{name: "missing audience", issuer: "https://hogedd.jp.auth0.com/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AUTH0_ISSUER_URL", tt.issuer)
			t.Setenv("AUTH0_AUDIENCE", tt.audience)

			if _, err := LoadAuthentication(); err == nil {
				t.Fatal("LoadAuthentication() error = nil, want error")
			}
		})
	}
}
