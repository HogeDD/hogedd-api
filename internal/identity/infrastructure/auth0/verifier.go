package auth0

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/auth0/go-jwt-middleware/v3/jwks"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/iwasawa/hogedd-api/internal/config"
	"github.com/iwasawa/hogedd-api/internal/identity"
)

const jwksRequestTimeout = 5 * time.Second

type tokenValidator interface {
	ValidateToken(context.Context, string) (any, error)
}

// Verifier はAuth0のJWKSを使い、RS256 Access Tokenを検証します。
type Verifier struct {
	validator tokenValidator
}

// NewVerifier はissuer、audience、JWKSを使うAccess Token検証器を構築します。
// JWKSは取得後にAuth0 SDK内でcacheされ、署名鍵のrotation時に更新されます。
func NewVerifier(cfg config.AuthenticationConfig) (*Verifier, error) {
	return newVerifier(cfg, &http.Client{Timeout: jwksRequestTimeout})
}

func newVerifier(cfg config.AuthenticationConfig, client *http.Client) (*Verifier, error) {
	issuerURL, err := url.Parse(cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("parse Auth0 issuer URL: %w", err)
	}

	provider, err := jwks.NewCachingProvider(
		jwks.WithIssuerURL(issuerURL),
		jwks.WithCustomClient(client),
	)
	if err != nil {
		return nil, fmt.Errorf("create Auth0 JWKS provider: %w", err)
	}

	tokenValidator, err := validator.New(
		validator.WithKeyFunc(provider.KeyFunc),
		validator.WithAlgorithm(validator.RS256),
		validator.WithIssuer(cfg.IssuerURL),
		validator.WithAudience(cfg.Audience),
	)
	if err != nil {
		return nil, fmt.Errorf("create Auth0 token validator: %w", err)
	}

	return &Verifier{validator: tokenValidator}, nil
}

// Verify はJWTの署名とregistered claimsを検証し、issuerとsubjectだけを返します。
func (v *Verifier) Verify(ctx context.Context, rawToken string) (identity.Identity, error) {
	claimsValue, err := v.validator.ValidateToken(ctx, rawToken)
	if err != nil {
		return identity.Identity{}, fmt.Errorf("validate Auth0 access token: %w", err)
	}

	claims, ok := claimsValue.(*validator.ValidatedClaims)
	if !ok {
		return identity.Identity{}, fmt.Errorf("validate Auth0 access token: unexpected claims type %T", claimsValue)
	}

	authenticated, err := identity.New(claims.RegisteredClaims.Issuer, claims.RegisteredClaims.Subject)
	if err != nil {
		return identity.Identity{}, fmt.Errorf("create identity from Auth0 claims: %w", err)
	}
	return authenticated, nil
}
