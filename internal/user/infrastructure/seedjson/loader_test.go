package seedjson

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsers(t *testing.T) {
	path := writeSeedFile(t, `{"version":1,"users":[{"auth_issuer":"https://hogedd.jp.auth0.com/","auth_subject":"auth0|owner","email":"owner@example.com","display_name":" Local Owner ","role":"owner","status":"active"}]}`)
	users, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(users) != 1 || users[0].DisplayName != "Local Owner" || users[0].Role.String() != "owner" {
		t.Fatalf("Load() = %#v", users)
	}
}

func TestLoadRejectsInvalidDocuments(t *testing.T) {
	tests := []struct{ name, body string }{
		{"unsupported version", `{"version":2,"users":[]}`},
		{"unknown field", `{"version":1,"unknown":true,"users":[]}`},
		{"invalid role", `{"version":1,"users":[{"auth_issuer":"https://example.com/","auth_subject":"auth0|x","email":"x@example.com","display_name":"X","role":"superuser","status":"active"}]}`},
		{"duplicate identity", `{"version":1,"users":[{"auth_issuer":"https://example.com/","auth_subject":"auth0|x","email":"x@example.com","display_name":"X","role":"owner","status":"active"},{"auth_issuer":"https://example.com/","auth_subject":"auth0|x","email":"y@example.com","display_name":"Y","role":"admin","status":"active"}]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Load(writeSeedFile(t, test.body)); err == nil {
				t.Fatal("Load() error = nil, want an error")
			}
		})
	}
}

func writeSeedFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "users.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
