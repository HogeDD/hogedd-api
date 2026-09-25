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
	// UsersFile は追加するLocal Auth0ユーザーを記述したJSONファイルです。
	UsersFile string
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

	usersFile := os.Getenv("SEED_USERS_FILE")
	if usersFile == "" {
		return LocalSeedConfig{}, fmt.Errorf("SEED_USERS_FILE is required")
	}

	return LocalSeedConfig{
		Database:  database,
		UsersFile: usersFile,
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
