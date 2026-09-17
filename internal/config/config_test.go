package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "18080")
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "debug")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Port != 18080 {
		t.Errorf("Port = %d, want %d", got.Port, 18080)
	}
	if got.Environment != "test" {
		t.Errorf("Environment = %q, want %q", got.Environment, "test")
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("PORT", "70000")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}

func TestLoadRuntimeUsesVercelEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("VERCEL_ENV", "preview")
	t.Setenv("LOG_LEVEL", "warn")

	got := LoadRuntime()
	if got.Environment != "preview" {
		t.Errorf("Environment = %q, want %q", got.Environment, "preview")
	}
	if got.LogLevel.String() != "WARN" {
		t.Errorf("LogLevel = %q, want %q", got.LogLevel, "WARN")
	}
}
