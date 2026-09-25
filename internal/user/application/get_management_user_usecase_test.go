package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iwasawa/hogedd-api/internal/identity"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

type managementUserFinderStub struct {
	user  *userdomain.User
	found bool
	err   error
}

func (s managementUserFinderStub) FindByIdentity(context.Context, string, string) (*userdomain.User, bool, error) {
	return s.user, s.found, s.err
}

func TestGetManagementUserUseCase(t *testing.T) {
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|user")
	now := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	user := func(role userdomain.Role, status userdomain.Status) *userdomain.User {
		result, err := userdomain.RestoreUser("user-1", authenticated.Issuer(), authenticated.Subject(), "user@example.com", true, role, status, now, now)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}

	tests := []struct {
		name    string
		finder  managementUserFinderStub
		wantErr error
	}{
		{name: "owner", finder: managementUserFinderStub{user: user(userdomain.RoleOwner, userdomain.StatusActive), found: true}},
		{name: "admin", finder: managementUserFinderStub{user: user(userdomain.RoleAdmin, userdomain.StatusActive), found: true}},
		{name: "member", finder: managementUserFinderStub{user: user(userdomain.RoleMember, userdomain.StatusActive), found: true}, wantErr: ErrManagementPermissionDenied},
		{name: "disabled", finder: managementUserFinderStub{user: user(userdomain.RoleOwner, userdomain.StatusDisabled), found: true}, wantErr: ErrManagementUserInactive},
		{name: "missing", finder: managementUserFinderStub{}, wantErr: ErrManagementUserNotFound},
		{name: "repository failure", finder: managementUserFinderStub{err: errors.New("database unavailable")}, wantErr: errors.New("database unavailable")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewGetManagementUserUseCase(tt.finder).Execute(context.Background(), authenticated)
			if tt.wantErr != nil {
				if err == nil || (tt.name != "repository failure" && !errors.Is(err, tt.wantErr)) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil || result.ID != "user-1" {
				t.Fatalf("result/error = %+v, %v", result, err)
			}
		})
	}
}
