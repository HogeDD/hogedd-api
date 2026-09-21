package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// AuthenticationConfig はAccess Tokenの発行者と対象APIを検証するための設定です。
type AuthenticationConfig struct {
	// IssuerURL は末尾のスラッシュを含むAuth0 tenantのissuerです。
	IssuerURL string
	// Audience はAccess Tokenの発行対象となるAPI識別子です。
	Audience string
}

// LoadAuthentication はAuth0のAccess Token検証設定を環境変数から読み取ります。
// issuerはHTTPSの絶対URL、audienceは空でないopaqueな識別子である必要があります。
func LoadAuthentication() (AuthenticationConfig, error) {
	issuer := os.Getenv("AUTH0_ISSUER_URL")
	if issuer == "" {
		return AuthenticationConfig{}, fmt.Errorf("AUTH0_ISSUER_URL is required")
	}

	issuerURL, err := url.Parse(issuer)
	if err != nil || issuerURL.Scheme != "https" || issuerURL.Host == "" {
		return AuthenticationConfig{}, fmt.Errorf("AUTH0_ISSUER_URL must be an absolute HTTPS URL")
	}
	if issuerURL.User != nil || issuerURL.RawQuery != "" || issuerURL.Fragment != "" {
		return AuthenticationConfig{}, fmt.Errorf("AUTH0_ISSUER_URL must not contain user info, query, or fragment")
	}
	if !strings.HasSuffix(issuerURL.Path, "/") {
		return AuthenticationConfig{}, fmt.Errorf("AUTH0_ISSUER_URL must end with a slash")
	}

	audience := os.Getenv("AUTH0_AUDIENCE")
	if audience == "" {
		return AuthenticationConfig{}, fmt.Errorf("AUTH0_AUDIENCE is required")
	}

	return AuthenticationConfig{
		IssuerURL: issuer,
		Audience:  audience,
	}, nil
}
