package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// DatabaseConfig はPostgreSQL接続poolの設定です。
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	PingTimeout     time.Duration
}

// LoadDatabase は実行時のPostgreSQL接続設定を環境変数から読み取ります。
func LoadDatabase() (DatabaseConfig, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_URL is required")
	}

	maxOpenConns, err := positiveIntEnv("DATABASE_MAX_OPEN_CONNS", 5)
	if err != nil {
		return DatabaseConfig{}, err
	}
	maxIdleConns, err := nonNegativeIntEnv("DATABASE_MAX_IDLE_CONNS", 2)
	if err != nil {
		return DatabaseConfig{}, err
	}
	if maxIdleConns > maxOpenConns {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_MAX_IDLE_CONNS must not exceed DATABASE_MAX_OPEN_CONNS")
	}

	return DatabaseConfig{
		URL:             url,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		PingTimeout:     5 * time.Second,
	}, nil
}

// LoadMigrationDatabaseURL はschema migration専用のPostgreSQL接続文字列を返します。
func LoadMigrationDatabaseURL() (string, error) {
	url := os.Getenv("DATABASE_URL_UNPOOLED")
	if override := os.Getenv("DATABASE_MIGRATION_URL"); override != "" {
		url = override
	}
	if url == "" {
		return "", fmt.Errorf("DATABASE_URL_UNPOOLED or DATABASE_MIGRATION_URL is required")
	}
	return url, nil
}

func positiveIntEnv(key string, fallback int) (int, error) {
	value, err := nonNegativeIntEnv(key, fallback)
	if err != nil {
		return 0, err
	}
	if value == 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func nonNegativeIntEnv(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	return value, nil
}
