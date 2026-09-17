package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// RuntimeConfig はローカルとVercel Functionsで共通して使用する実行時設定です。
type RuntimeConfig struct {
	// Environment はdevelopmentやproductionなどの実行環境名です。
	Environment string
	// LogLevel は構造化ログへ出力する最低レベルです。
	LogLevel slog.Level
}

// Config はローカルHTTPサーバのプロセス起動時に確定する設定を表します。
type Config struct {
	RuntimeConfig
	// Port はローカルHTTPサーバが待ち受けるTCPポートです。
	Port int
	// ShutdownTimeout はgraceful shutdownの完了を待つ最大時間です。
	ShutdownTimeout time.Duration
	// ReadTimeout はリクエストヘッダーの読み取りを待つ最大時間です。
	ReadTimeout time.Duration
	// WriteTimeout はHTTPレスポンスの書き込みを待つ最大時間です。
	WriteTimeout time.Duration
	// IdleTimeout はkeep-alive接続の次のリクエストを待つ最大時間です。
	IdleTimeout time.Duration
}

// LoadRuntime は実行環境に依存しない共通設定を環境変数から読み取ります。
// APP_ENVが未指定の場合はVercel提供のVERCEL_ENVを参照します。
func LoadRuntime() RuntimeConfig {
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = envOrDefault("VERCEL_ENV", "development")
	}

	return RuntimeConfig{
		Environment: environment,
		LogLevel:    parseLogLevel(envOrDefault("LOG_LEVEL", "info")),
	}
}

// Load は環境変数を読み取り、検証済みのConfigを返します。
// 不正な設定値がある場合は、デフォルト値へ暗黙に置き換えずエラーを返します。
func Load() (Config, error) {
	config := Config{
		RuntimeConfig:   LoadRuntime(),
		Port:            8080,
		ShutdownTimeout: 10 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
	}

	if rawPort := os.Getenv("PORT"); rawPort != "" {
		port, err := strconv.Atoi(rawPort)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("PORT must be an integer between 1 and 65535")
		}
		config.Port = port
	}

	return config, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseLogLevel(value string) slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return slog.LevelInfo
	}
	return level
}
