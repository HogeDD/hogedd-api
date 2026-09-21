package auth0

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iwasawa/hogedd-api/internal/config"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

func TestVerifierValidatesAuth0AccessToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	privateJWK, err := jwk.Import(privateKey)
	if err != nil {
		t.Fatalf("import private JWK: %v", err)
	}
	if err := privateJWK.Set(jwk.KeyIDKey, "test-key"); err != nil {
		t.Fatalf("set private JWK key ID: %v", err)
	}
	publicJWK, err := jwk.Import(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("import public JWK: %v", err)
	}
	if err := publicJWK.Set(jwk.KeyIDKey, "test-key"); err != nil {
		t.Fatalf("set public JWK key ID: %v", err)
	}
	keySet := jwk.NewSet()
	if err := keySet.AddKey(publicJWK); err != nil {
		t.Fatalf("add public JWK: %v", err)
	}

	var issuer string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":   issuer,
				"jwks_uri": issuer + "jwks.json",
			})
		case "/jwks.json":
			_ = json.NewEncoder(w).Encode(keySet)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	issuer = server.URL + "/"

	verifier, err := newVerifier(config.AuthenticationConfig{
		IssuerURL: issuer,
		Audience:  "https://api.hogedd.com",
	}, server.Client())
	if err != nil {
		t.Fatalf("newVerifier() error = %v", err)
	}

	tests := []struct {
		name     string
		issuer   string
		audience string
		subject  string
		expires  time.Time
		wantErr  bool
	}{
		{
			name:     "valid",
			issuer:   issuer,
			audience: "https://api.hogedd.com",
			subject:  "auth0|owner",
			expires:  time.Now().Add(time.Hour),
		},
		{
			name:     "wrong issuer",
			issuer:   "https://attacker.example/",
			audience: "https://api.hogedd.com",
			subject:  "auth0|owner",
			expires:  time.Now().Add(time.Hour),
			wantErr:  true,
		},
		{
			name:     "wrong audience",
			issuer:   issuer,
			audience: "https://other-api.example",
			subject:  "auth0|owner",
			expires:  time.Now().Add(time.Hour),
			wantErr:  true,
		},
		{
			name:     "expired",
			issuer:   issuer,
			audience: "https://api.hogedd.com",
			subject:  "auth0|owner",
			expires:  time.Now().Add(-time.Hour),
			wantErr:  true,
		},
		{
			name:     "missing subject",
			issuer:   issuer,
			audience: "https://api.hogedd.com",
			expires:  time.Now().Add(time.Hour),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwt.NewBuilder().
				Issuer(tt.issuer).
				Audience([]string{tt.audience}).
				Subject(tt.subject).
				Expiration(tt.expires).
				Build()
			if err != nil {
				t.Fatalf("build token: %v", err)
			}
			rawToken, err := jwt.Sign(token, jwt.WithKey(jwa.RS256(), privateJWK))
			if err != nil {
				t.Fatalf("sign token: %v", err)
			}

			got, err := verifier.Verify(context.Background(), string(rawToken))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && (got.Issuer() != tt.issuer || got.Subject() != tt.subject) {
				t.Errorf("Identity = (%q, %q)", got.Issuer(), got.Subject())
			}
		})
	}
}
