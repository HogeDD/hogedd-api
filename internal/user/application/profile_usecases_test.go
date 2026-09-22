package application

import (
	"context"
	"errors"
	"testing"

	"github.com/iwasawa/hogedd-api/internal/identity"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

type profileRepositoryStub struct {
	profile *userdomain.Profile
	status  userdomain.Status
	found   bool
	err     error
}

func (s *profileRepositoryStub) FindByIdentity(context.Context, string, string) (*userdomain.Profile, userdomain.Status, bool, error) {
	return s.profile, s.status, s.found, s.err
}

func (s *profileRepositoryStub) SaveByIdentity(_ context.Context, _, _, displayName string) (*userdomain.Profile, userdomain.Status, error) {
	if s.err != nil {
		return nil, "", s.err
	}
	profile, err := userdomain.NewProfile("user-1", displayName)
	return profile, s.status, err
}

func testIdentity(t *testing.T) identity.Identity {
	t.Helper()
	value, err := identity.New("https://issuer.example/", "subject")
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestGetCurrentProfileUseCase(t *testing.T) {
	profile, _ := userdomain.NewProfile("user-1", "HogeDD")
	tests := []struct {
		name       string
		repository *profileRepositoryStub
		want       error
	}{
		{name: "profile", repository: &profileRepositoryStub{profile: profile, status: userdomain.StatusActive, found: true}},
		{name: "missing user", repository: &profileRepositoryStub{}, want: ErrUserNotFound},
		{name: "missing profile", repository: &profileRepositoryStub{status: userdomain.StatusActive}, want: ErrProfileNotFound},
		{name: "disabled", repository: &profileRepositoryStub{profile: profile, status: userdomain.StatusDisabled, found: true}, want: ErrUserDisabled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewGetCurrentProfileUseCase(tt.repository).Execute(context.Background(), testIdentity(t))
			if !errors.Is(err, tt.want) {
				t.Fatalf("Execute() error = %v, want %v", err, tt.want)
			}
			if tt.want == nil && result.DisplayName != "HogeDD" {
				t.Fatalf("DisplayName = %q", result.DisplayName)
			}
		})
	}
}

func TestUpdateCurrentProfileUseCase(t *testing.T) {
	repository := &profileRepositoryStub{status: userdomain.StatusActive}
	result, err := NewUpdateCurrentProfileUseCase(repository).Execute(context.Background(), testIdentity(t), "  HogeDD   Owner ")
	if err != nil {
		t.Fatal(err)
	}
	if result.DisplayName != "HogeDD Owner" {
		t.Fatalf("DisplayName = %q", result.DisplayName)
	}
}

func TestUpdateCurrentProfileUseCaseRejectsInvalidName(t *testing.T) {
	_, err := NewUpdateCurrentProfileUseCase(&profileRepositoryStub{status: userdomain.StatusActive}).Execute(context.Background(), testIdentity(t), " ")
	if !errors.Is(err, userdomain.ErrDisplayNameRequired) {
		t.Fatalf("Execute() error = %v", err)
	}
}
