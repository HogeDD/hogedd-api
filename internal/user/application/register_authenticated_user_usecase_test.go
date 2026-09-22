package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iwasawa/hogedd-api/internal/identity"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

type profileProviderStub struct {
	profile AuthenticatedProfile
	err     error
}

func (s profileProviderStub) Fetch(context.Context, string) (AuthenticatedProfile, error) {
	return s.profile, s.err
}

type registrarStub struct {
	params  RegisterUserParams
	user    *userdomain.User
	created bool
	err     error
}

func (s *registrarStub) Register(_ context.Context, params RegisterUserParams) (*userdomain.User, bool, error) {
	s.params = params
	return s.user, s.created, s.err
}

func TestRegisterAuthenticatedUserCreatesMember(t *testing.T) {
	authenticated, err := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	registered, err := userdomain.RestoreUser(
		"0199-user", authenticated.Issuer(), authenticated.Subject(), "owner@example.com", true,
		userdomain.RoleMember, userdomain.StatusActive, now, now,
	)
	if err != nil {
		t.Fatal(err)
	}
	registrar := &registrarStub{user: registered, created: true}
	useCase := NewRegisterAuthenticatedUserUseCase(profileProviderStub{profile: AuthenticatedProfile{
		Subject:       authenticated.Subject(),
		Email:         "owner@example.com",
		EmailVerified: true,
	}}, registrar)

	result, err := useCase.Execute(context.Background(), authenticated, "access-token")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.Created || result.Role != "member" || result.Status != "active" {
		t.Errorf("Execute() result = %+v", result)
	}
	if registrar.params.InitialRole != userdomain.RoleMember {
		t.Errorf("InitialRole = %q, want member", registrar.params.InitialRole)
	}
	if registrar.params.AuthIssuer != authenticated.Issuer() || registrar.params.AuthSubject != authenticated.Subject() {
		t.Errorf("identity params = %+v", registrar.params)
	}
}

func TestRegisterAuthenticatedUserPreservesExistingAuthorization(t *testing.T) {
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	now := time.Now()
	existing, _ := userdomain.RestoreUser(
		"0199-user", authenticated.Issuer(), authenticated.Subject(), "new@example.com", true,
		userdomain.RoleOwner, userdomain.StatusDisabled, now, now,
	)
	registrar := &registrarStub{user: existing, created: false}
	useCase := NewRegisterAuthenticatedUserUseCase(profileProviderStub{profile: AuthenticatedProfile{
		Subject: authenticated.Subject(), Email: "new@example.com", EmailVerified: true,
	}}, registrar)

	result, err := useCase.Execute(context.Background(), authenticated, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	if result.Created || result.Role != "owner" || result.Status != "disabled" {
		t.Errorf("Execute() result = %+v", result)
	}
}

func TestRegisterAuthenticatedUserRejectsMismatchedProfile(t *testing.T) {
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	useCase := NewRegisterAuthenticatedUserUseCase(profileProviderStub{profile: AuthenticatedProfile{
		Subject: "auth0|attacker", Email: "attacker@example.com",
	}}, &registrarStub{})

	_, err := useCase.Execute(context.Background(), authenticated, "access-token")
	if !errors.Is(err, ErrProfileIdentityMismatch) {
		t.Fatalf("Execute() error = %v, want identity mismatch", err)
	}
}

func TestRegisterAuthenticatedUserRequiresAccessToken(t *testing.T) {
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	useCase := NewRegisterAuthenticatedUserUseCase(profileProviderStub{}, &registrarStub{})

	_, err := useCase.Execute(context.Background(), authenticated, "")
	if !errors.Is(err, ErrAccessTokenRequired) {
		t.Fatalf("Execute() error = %v, want access token required", err)
	}
}
