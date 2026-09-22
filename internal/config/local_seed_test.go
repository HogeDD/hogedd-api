package config

import "testing"

func TestLoadLocalSeedAcceptsDockerDatabase(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgresql://hogedd:hogedd@db:5432/hogedd?sslmode=disable")
	t.Setenv("SEED_AUTH_ISSUER", "https://local.example.invalid/")
	t.Setenv("SEED_AUTH_SUBJECT", "auth0|local-owner")
	t.Setenv("SEED_AUTH_EMAIL", "owner@example.invalid")

	got, err := LoadLocalSeed()
	if err != nil {
		t.Fatalf("LoadLocalSeed() error = %v", err)
	}
	if !got.IncludeIdentity || got.DisplayName != "Local Owner" {
		t.Fatalf("LoadLocalSeed() = %#v", got)
	}
}

func TestLoadLocalSeedRejectsProductionEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgresql://hogedd:hogedd@db:5432/hogedd")

	if _, err := LoadLocalSeed(); err == nil {
		t.Fatal("LoadLocalSeed() error = nil, want an error")
	}
}

func TestLoadLocalSeedRejectsRemoteDatabase(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgresql://user:password@example.neon.tech/hogedd")

	if _, err := LoadLocalSeed(); err == nil {
		t.Fatal("LoadLocalSeed() error = nil, want an error")
	}
}

func TestLoadLocalSeedRejectsPartialIdentity(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgresql://hogedd:hogedd@localhost:5432/hogedd")
	t.Setenv("SEED_AUTH_SUBJECT", "auth0|local-owner")

	if _, err := LoadLocalSeed(); err == nil {
		t.Fatal("LoadLocalSeed() error = nil, want an error")
	}
}
