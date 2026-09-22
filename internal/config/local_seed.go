package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

// LocalSeedConfig はLocal DBへseedを投入するための設定です。
type LocalSeedConfig struct {
	// Database は接続先をLocalへ制限したPostgreSQL設定です。
	Database DatabaseConfig
	// AuthIssuer は任意で追加するLocal Auth0ユーザーのissuerです。
	AuthIssuer string
	// AuthSubject は任意で追加するLocal Auth0ユーザーのsubjectです。
	AuthSubject string
	// Email は任意で追加するLocal Auth0ユーザーのemailです。
	Email string
	// DisplayName は任意で追加するLocal Auth0ユーザーの表示名です。
	DisplayName string
	// IncludeIdentity はLocal Auth0ユーザーをseedへ含めるかを表します。
	IncludeIdentity bool
}

// LoadLocalSeed はLocal専用のseed設定を読み取り、Production接続を拒否します。
func LoadLocalSeed() (LocalSeedConfig, error) {
	if os.Getenv("APP_ENV") != "development" {
		return LocalSeedConfig{}, fmt.Errorf("APP_ENV must be development to run local seed")
	}
	database, err := LoadDatabase()
	if err != nil {
		return LocalSeedConfig{}, err
	}
	if err := validateLocalDatabaseURL(database.URL); err != nil {
		return LocalSeedConfig{}, err
	}

	authIssuer := os.Getenv("SEED_AUTH_ISSUER")
	authSubject := os.Getenv("SEED_AUTH_SUBJECT")
	email := os.Getenv("SEED_AUTH_EMAIL")
	provided := 0
	for _, value := range []string{authIssuer, authSubject, email} {
		if value != "" {
			provided++
		}
	}
	if provided != 0 && provided != 3 {
		return LocalSeedConfig{}, fmt.Errorf("SEED_AUTH_ISSUER, SEED_AUTH_SUBJECT and SEED_AUTH_EMAIL must be set together")
	}
	displayName := os.Getenv("SEED_AUTH_DISPLAY_NAME")
	if displayName == "" {
		displayName = "Local Owner"
	}

	return LocalSeedConfig{
		Database:        database,
		AuthIssuer:      authIssuer,
		AuthSubject:     authSubject,
		Email:           email,
		DisplayName:     displayName,
		IncludeIdentity: provided == 3,
	}, nil
}

func validateLocalDatabaseURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL for local seed: %w", err)
	}
	host := parsed.Hostname()
	if host == "db" || host == "localhost" {
		return nil
	}
	if address := net.ParseIP(host); address != nil && address.IsLoopback() {
		return nil
	}
	return fmt.Errorf("local seed refuses non-local database host %q", host)
}
