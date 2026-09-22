package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func TestProfileRepositoryFindByIdentity(t *testing.T) {
	repository := &ProfileRepository{queryRow: func(context.Context, string, ...any) rowScanner {
		return scannerStub{values: []any{"0199-user", "active", "HogeDD"}}
	}}
	profile, status, found, err := repository.FindByIdentity(context.Background(), "issuer", "subject")
	if err != nil || !found || status.String() != "active" || profile.DisplayName() != "HogeDD" {
		t.Fatalf("FindByIdentity() = %#v, %q, %v, %v", profile, status, found, err)
	}
}

func TestProfileRepositoryFindsUserWithoutProfile(t *testing.T) {
	repository := &ProfileRepository{queryRow: func(context.Context, string, ...any) rowScanner {
		return scannerStub{values: []any{"0199-user", "active", sql.NullString{}}}
	}}
	profile, _, found, err := repository.FindByIdentity(context.Background(), "issuer", "subject")
	if err != nil || found || profile != nil {
		t.Fatalf("FindByIdentity() = %#v, %v, %v", profile, found, err)
	}
}

func TestProfileRepositorySavesDisplayName(t *testing.T) {
	repository := &ProfileRepository{queryRow: func(context.Context, string, ...any) rowScanner {
		return scannerStub{values: []any{"0199-user", "active", "HogeDD"}}
	}}
	profile, status, err := repository.SaveByIdentity(context.Background(), "issuer", "subject", "HogeDD")
	if err != nil || status.String() != "active" || profile.DisplayName() != "HogeDD" {
		t.Fatalf("SaveByIdentity() = %#v, %q, %v", profile, status, err)
	}
}
