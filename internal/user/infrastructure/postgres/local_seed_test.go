package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

type seedExecutorStub struct {
	queries []string
	args    [][]any
}

func (s *seedExecutorStub) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	s.queries = append(s.queries, query)
	s.args = append(s.args, args)
	return nil, nil
}

func TestSeedLocalUsersIncludesRepresentativeStates(t *testing.T) {
	executor := &seedExecutorStub{}
	users := defaultLocalSeedUsers()

	if err := seedLocalUsers(context.Background(), executor, users); err != nil {
		t.Fatalf("seedLocalUsers() error = %v", err)
	}
	if len(executor.queries) != 8 {
		t.Fatalf("query count = %d, want 8", len(executor.queries))
	}

	wantRoleStatus := [][2]string{
		{"owner", "active"},
		{"admin", "active"},
		{"member", "active"},
		{"member", "disabled"},
	}
	for index, want := range wantRoleStatus {
		args := executor.args[index*2]
		if args[5] != want[0] || args[6] != want[1] {
			t.Errorf("user %d role/status = %v/%v, want %v/%v", index, args[5], args[6], want[0], want[1])
		}
	}
	if !strings.Contains(executor.queries[7], "DELETE FROM user_profiles") {
		t.Errorf("disabled member profile query = %q", executor.queries[7])
	}
}

func TestSeedLocalUsersUsesParametersAndUpsert(t *testing.T) {
	executor := &seedExecutorStub{}
	users := defaultLocalSeedUsers()

	if err := seedLocalUsers(context.Background(), executor, users[:1]); err != nil {
		t.Fatalf("seedLocalUsers() error = %v", err)
	}
	if !strings.Contains(executor.queries[0], "ON CONFLICT") || strings.Contains(executor.queries[0], users[0].email) {
		t.Errorf("user query is not a parameterized upsert: %q", executor.queries[0])
	}
	if !strings.Contains(executor.queries[1], "ON CONFLICT") || strings.Contains(executor.queries[1], users[0].displayName) {
		t.Errorf("profile query is not a parameterized upsert: %q", executor.queries[1])
	}
}

func TestSeedLocalUsersLetsDatabaseGenerateIdentityID(t *testing.T) {
	executor := &seedExecutorStub{}
	identity := localSeedUser{
		authIssuer:    "https://local.example.invalid/",
		authSubject:   "auth0|local-owner",
		email:         "owner@example.invalid",
		emailVerified: true,
		role:          userdomain.RoleOwner,
		status:        userdomain.StatusActive,
		displayName:   "Local Owner",
	}

	if err := seedLocalUsers(context.Background(), executor, []localSeedUser{identity}); err != nil {
		t.Fatalf("seedLocalUsers() error = %v", err)
	}
	if strings.Contains(executor.queries[0], "INSERT INTO users (id,") {
		t.Errorf("identity query fixes a reusable UUID: %q", executor.queries[0])
	}
	if len(executor.args[0]) != 6 {
		t.Errorf("identity args = %v, want 6 values", executor.args[0])
	}
}
