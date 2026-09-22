package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iwasawa/hogedd-api/internal/identity"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

type userFinderStub struct {
	user  *userdomain.User
	found bool
	err   error
}

func (s userFinderStub) FindByIdentity(context.Context, string, string) (*userdomain.User, bool, error) {
	return s.user, s.found, s.err
}

func TestGetCurrentUser(t *testing.T) {
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	user, _ := userdomain.RestoreUser(
		"0199-user", authenticated.Issuer(), authenticated.Subject(), "owner@example.com", true,
		userdomain.RoleMember, userdomain.StatusActive, now, now,
	)
	useCase := NewGetCurrentUserUseCase(userFinderStub{user: user, found: true})

	result, err := useCase.Execute(context.Background(), authenticated)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "0199-user" || result.Email != "owner@example.com" || result.Created {
		t.Errorf("Execute() = %+v", result)
	}
}

func TestGetCurrentUserReturnsNotFound(t *testing.T) {
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|missing")
	useCase := NewGetCurrentUserUseCase(userFinderStub{})

	_, err := useCase.Execute(context.Background(), authenticated)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("Execute() error = %v", err)
	}
}
