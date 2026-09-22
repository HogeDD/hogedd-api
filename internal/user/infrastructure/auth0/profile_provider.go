package auth0

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
)

const (
	profileRequestTimeout = 5 * time.Second
	maxProfileBodyBytes   = 1 << 20
)

// ProfileProvider はAuth0 /userinfoから認証主体の連絡先snapshotを取得します。
type ProfileProvider struct {
	endpoint *url.URL
	client   *http.Client
}

// NewProfileProvider はissuerとtimeout付きHTTP clientからProfileProviderを構築します。
func NewProfileProvider(issuer string) (*ProfileProvider, error) {
	return newProfileProvider(issuer, &http.Client{Timeout: profileRequestTimeout})
}

func newProfileProvider(issuer string, client *http.Client) (*ProfileProvider, error) {
	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return nil, fmt.Errorf("parse Auth0 issuer URL: %w", err)
	}
	endpoint, err := issuerURL.Parse("userinfo")
	if err != nil {
		return nil, fmt.Errorf("build Auth0 userinfo URL: %w", err)
	}
	return &ProfileProvider{endpoint: endpoint, client: client}, nil
}

// Fetch はuser-issued Access TokenをAuth0 /userinfoへ送り、subjectとemailを取得します。
func (p *ProfileProvider) Fetch(ctx context.Context, accessToken string) (userapp.AuthenticatedProfile, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint.String(), nil)
	if err != nil {
		return userapp.AuthenticatedProfile{}, fmt.Errorf("create Auth0 userinfo request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+accessToken)

	response, err := p.client.Do(request)
	if err != nil {
		return userapp.AuthenticatedProfile{}, fmt.Errorf("request Auth0 userinfo: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		io.Copy(io.Discard, io.LimitReader(response.Body, maxProfileBodyBytes))
		return userapp.AuthenticatedProfile{}, fmt.Errorf("Auth0 userinfo returned status %d", response.StatusCode)
	}

	var body struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxProfileBodyBytes))
	if err := decoder.Decode(&body); err != nil {
		return userapp.AuthenticatedProfile{}, fmt.Errorf("decode Auth0 userinfo response: %w", err)
	}
	if body.Subject == "" || body.Email == "" {
		return userapp.AuthenticatedProfile{}, fmt.Errorf("Auth0 userinfo response is missing required claims")
	}

	return userapp.AuthenticatedProfile{
		Subject:       body.Subject,
		Email:         body.Email,
		EmailVerified: body.EmailVerified,
	}, nil
}
