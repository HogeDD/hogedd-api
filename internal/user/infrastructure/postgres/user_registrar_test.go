package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

type scannerStub struct {
	err    error
	values []any
}

func (s scannerStub) Scan(destinations ...any) error {
	if s.err != nil {
		return s.err
	}
	for index, value := range s.values {
		switch destination := destinations[index].(type) {
		case *string:
			*destination = value.(string)
		case *bool:
			*destination = value.(bool)
		case *sql.NullTime:
			*destination = sql.NullTime{Time: value.(time.Time), Valid: true}
		}
	}
	return nil
}

func userRow(role, status string) scannerStub {
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	return scannerStub{values: []any{
		"0199-user",
		"https://hogedd.jp.auth0.com/",
		"auth0|owner",
		"owner@example.com",
		true,
		role,
		status,
		now,
		now,
	}}
}

func TestUserRegistrarInsertsNewMember(t *testing.T) {
	var queries []string
	registrar := &UserRegistrar{queryRow: func(_ context.Context, query string, args ...any) rowScanner {
		queries = append(queries, query)
		if got := args[4]; got != "member" {
			t.Errorf("initial role = %v", got)
		}
		return userRow("member", "active")
	}}

	registered, created, err := registrar.Register(context.Background(), userapp.RegisterUserParams{
		AuthIssuer:    "https://hogedd.jp.auth0.com/",
		AuthSubject:   "auth0|owner",
		Email:         "owner@example.com",
		EmailVerified: true,
		InitialRole:   userdomain.RoleMember,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if !created || registered.Role() != userdomain.RoleMember {
		t.Errorf("Register() = created %v, role %q", created, registered.Role())
	}
	if len(queries) != 1 || !strings.Contains(queries[0], "INSERT INTO users") {
		t.Errorf("queries = %v", queries)
	}
}

func TestUserRegistrarUpdatesOnlyExistingContactSnapshot(t *testing.T) {
	var queries []string
	registrar := &UserRegistrar{queryRow: func(_ context.Context, query string, _ ...any) rowScanner {
		queries = append(queries, query)
		if len(queries) == 1 {
			return scannerStub{err: sql.ErrNoRows}
		}
		return userRow("owner", "disabled")
	}}

	registered, created, err := registrar.Register(context.Background(), userapp.RegisterUserParams{
		AuthIssuer:    "https://hogedd.jp.auth0.com/",
		AuthSubject:   "auth0|owner",
		Email:         "owner@example.com",
		EmailVerified: true,
		InitialRole:   userdomain.RoleMember,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if created || registered.Role() != userdomain.RoleOwner || registered.Status() != userdomain.StatusDisabled {
		t.Errorf("Register() = created %v, role/status %q/%q", created, registered.Role(), registered.Status())
	}
	if len(queries) != 2 || !strings.Contains(queries[1], "SET email = $3") {
		t.Errorf("queries = %v", queries)
	}
	if strings.Contains(queries[1], "role =") || strings.Contains(queries[1], "status =") {
		t.Errorf("update query changes authorization: %s", queries[1])
	}
}

func TestUserRegistrarFindsByIdentity(t *testing.T) {
	registrar := &UserRegistrar{queryRow: func(_ context.Context, query string, args ...any) rowScanner {
		if !strings.Contains(query, "WHERE auth_issuer = $1 AND auth_subject = $2") {
			t.Errorf("query = %s", query)
		}
		if args[0] != "https://hogedd.jp.auth0.com/" || args[1] != "auth0|owner" {
			t.Errorf("args = %v", args)
		}
		return userRow("member", "active")
	}}

	user, found, err := registrar.FindByIdentity(
		context.Background(), "https://hogedd.jp.auth0.com/", "auth0|owner",
	)
	if err != nil || !found || user.ID() != "0199-user" {
		t.Fatalf("FindByIdentity() = %v, %v, %v", user, found, err)
	}
}

func TestUserRegistrarReturnsNotFound(t *testing.T) {
	registrar := &UserRegistrar{queryRow: func(context.Context, string, ...any) rowScanner {
		return scannerStub{err: sql.ErrNoRows}
	}}

	user, found, err := registrar.FindByIdentity(context.Background(), "issuer", "subject")
	if err != nil || found || user != nil {
		t.Fatalf("FindByIdentity() = %v, %v, %v", user, found, err)
	}
}
