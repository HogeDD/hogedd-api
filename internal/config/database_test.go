package config

import "testing"

func TestLoadDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://example.invalid/hogedd")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "8")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "3")

	got, err := LoadDatabase()
	if err != nil {
		t.Fatalf("LoadDatabase() error = %v", err)
	}
	if got.URL != "postgresql://example.invalid/hogedd" {
		t.Errorf("URL = %q", got.URL)
	}
	if got.MaxOpenConns != 8 || got.MaxIdleConns != 3 {
		t.Errorf("pool size = (%d, %d)", got.MaxOpenConns, got.MaxIdleConns)
	}
}

func TestLoadDatabaseRequiresURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := LoadDatabase(); err == nil {
		t.Fatal("LoadDatabase() error = nil, want an error")
	}
}

func TestLoadDatabaseRejectsInvalidPoolSize(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://example.invalid/hogedd")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "2")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "3")

	if _, err := LoadDatabase(); err == nil {
		t.Fatal("LoadDatabase() error = nil, want an error")
	}
}

func TestLoadMigrationDatabaseURLUsesNeonUnpooledURL(t *testing.T) {
	t.Setenv("DATABASE_URL_UNPOOLED", "postgresql://direct.example.invalid/hogedd")
	t.Setenv("DATABASE_MIGRATION_URL", "")

	got, err := LoadMigrationDatabaseURL()
	if err != nil {
		t.Fatalf("LoadMigrationDatabaseURL() error = %v", err)
	}
	if got != "postgresql://direct.example.invalid/hogedd" {
		t.Errorf("URL = %q", got)
	}
}

func TestLoadMigrationDatabaseURLPrefersExplicitOverride(t *testing.T) {
	t.Setenv("DATABASE_URL_UNPOOLED", "postgresql://neon.example.invalid/hogedd")
	t.Setenv("DATABASE_MIGRATION_URL", "postgresql://override.example.invalid/hogedd")

	got, err := LoadMigrationDatabaseURL()
	if err != nil {
		t.Fatalf("LoadMigrationDatabaseURL() error = %v", err)
	}
	if got != "postgresql://override.example.invalid/hogedd" {
		t.Errorf("URL = %q", got)
	}
}
