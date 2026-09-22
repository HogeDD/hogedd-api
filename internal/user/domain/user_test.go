package domain

import (
	"errors"
	"testing"
	"time"
)

func TestRestoreUser(t *testing.T) {
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	user, err := RestoreUser(
		"0199-user",
		"https://hogedd.jp.auth0.com/",
		"auth0|owner",
		" Owner@Example.com ",
		true,
		RoleOwner,
		StatusActive,
		now,
		now,
	)
	if err != nil {
		t.Fatalf("RestoreUser() error = %v", err)
	}
	if user.Email() != "owner@example.com" {
		t.Errorf("Email() = %q", user.Email())
	}
	if user.Role() != RoleOwner || user.Status() != StatusActive {
		t.Errorf("role/status = %q/%q", user.Role(), user.Status())
	}
}

func TestRestoreUserRejectsInvalidValues(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		id      string
		issuer  string
		subject string
		email   string
		role    Role
		status  Status
		wantErr error
	}{
		{name: "missing ID", issuer: "issuer", subject: "subject", email: "a@example.com", role: RoleMember, status: StatusActive, wantErr: ErrUserIDRequired},
		{name: "missing issuer", id: "id", subject: "subject", email: "a@example.com", role: RoleMember, status: StatusActive, wantErr: ErrAuthIssuerRequired},
		{name: "missing subject", id: "id", issuer: "issuer", email: "a@example.com", role: RoleMember, status: StatusActive, wantErr: ErrAuthSubjectRequired},
		{name: "invalid email", id: "id", issuer: "issuer", subject: "subject", email: "invalid", role: RoleMember, status: StatusActive, wantErr: ErrInvalidEmail},
		{name: "invalid role", id: "id", issuer: "issuer", subject: "subject", email: "a@example.com", role: Role("unknown"), status: StatusActive},
		{name: "invalid status", id: "id", issuer: "issuer", subject: "subject", email: "a@example.com", role: RoleMember, status: Status("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RestoreUser(tt.id, tt.issuer, tt.subject, tt.email, false, tt.role, tt.status, now, now)
			if err == nil {
				t.Fatal("RestoreUser() error = nil")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("RestoreUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoleAndStatusParsers(t *testing.T) {
	for _, role := range []Role{RoleOwner, RoleAdmin, RoleMember} {
		if got, err := ParseRole(role.String()); err != nil || got != role {
			t.Errorf("ParseRole(%q) = %q, %v", role, got, err)
		}
	}
	for _, status := range []Status{StatusActive, StatusDisabled} {
		if got, err := ParseStatus(status.String()); err != nil || got != status {
			t.Errorf("ParseStatus(%q) = %q, %v", status, got, err)
		}
	}
}
