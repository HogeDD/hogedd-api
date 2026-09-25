package config

import "testing"

func TestLoadLocalSeedAcceptsDockerDatabase(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgresql://hogedd:hogedd@db:5432/hogedd?sslmode=disable")
	t.Setenv("SEED_USERS_FILE", "/seed/users.json")

	got, err := LoadLocalSeed()
	if err != nil {
		t.Fatalf("LoadLocalSeed() error = %v", err)
	}
	if got.UsersFile != "/seed/users.json" {
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

func TestLoadLocalSeedRequiresUsersFile(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgresql://hogedd:hogedd@localhost:5432/hogedd")

	if _, err := LoadLocalSeed(); err == nil {
		t.Fatal("LoadLocalSeed() error = nil, want an error")
	}
}
