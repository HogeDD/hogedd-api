package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/iwasawa/hogedd-api/internal/config"
)

// Server はローカル実行用HTTPサーバの起動と終了処理を管理します。
type Server struct {
	server          *http.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

// New は設定済みのhttp.Serverを構築します。
// サーバの起動は行わないため、呼び出し側はRunを実行する必要があります。
func New(cfg config.Config, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           handler,
			ReadHeaderTimeout: cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
		logger:          logger,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

// Run はHTTPサーバを起動し、ctxの終了時にgraceful shutdownを行います。
// 待ち受けに失敗した場合、原因をラップしたエラーを返します。
func (s *Server) Run(ctx context.Context) error {
	shutdownComplete := make(chan struct{})
	go func() {
		defer close(shutdownComplete)
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	s.logger.Info("HTTP server started", "address", s.server.Addr)
	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}

	<-shutdownComplete
	return nil
}
