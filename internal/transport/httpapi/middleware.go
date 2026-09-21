package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"
)

// Middleware はHTTPハンドラーへ横断的な処理を追加する関数です。
type Middleware func(http.Handler) http.Handler

// MiddlewareStack は全endpointへ適用する共通middlewareを保持します。
// Wrapへ追加のmiddlewareを渡すと、そのendpointだけに適用できます。
type MiddlewareStack struct {
	common []Middleware
}

type contextKey string

const requestIDKey contextKey = "request_id"

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// Chain はmiddlewareを宣言順に適用したHTTPハンドラーを返します。
// 最初に指定したMiddlewareがリクエストを最初に受け取ります。
func Chain(handler http.Handler, middleware ...Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// NewMiddlewareStack は宣言順に実行される共通middlewareを保持するStackを構築します。
func NewMiddlewareStack(common ...Middleware) *MiddlewareStack {
	return &MiddlewareStack{common: append([]Middleware(nil), common...)}
}

// Wrap は共通middlewareの内側へendpoint固有middlewareを追加してHandlerを包みます。
func (s *MiddlewareStack) Wrap(handler http.Handler, endpoint ...Middleware) http.Handler {
	middleware := make([]Middleware, 0, len(s.common)+len(endpoint))
	middleware = append(middleware, s.common...)
	middleware = append(middleware, endpoint...)
	return Chain(handler, middleware...)
}

// RequestIDFromContext はリクエストコンテキストに保存されたrequest IDを返します。
// request IDが設定されていない場合は空文字列を返します。
func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}

// RequestID はrequest IDを検証または生成し、コンテキストとレスポンスへ設定します。
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if !validRequestID.MatchString(requestID) {
				requestID = newRequestID()
			}

			w.Header().Set("X-Request-ID", requestID)
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SecurityHeaders はAPIレスポンスへ共通のセキュリティヘッダーを付与します。
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			next.ServeHTTP(w, r)
		})
	}
}

// AccessLog は完了したHTTPリクエストを構造化ログとして記録します。
func AccessLog(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			observer := &responseObserver{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(observer, r)
			logger.InfoContext(r.Context(), "HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"request_id", RequestIDFromContext(r.Context()),
				"status", observer.status,
				"response_bytes", observer.bytes,
				"duration_ms", time.Since(startedAt).Milliseconds(),
			)
		})
	}
}

type responseObserver struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (w *responseObserver) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseObserver) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	written, err := w.ResponseWriter.Write(body)
	w.bytes += written
	return written, err
}

func (w *responseObserver) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Recover はpanicを捕捉し、内部情報を露出しない共通エラーレスポンスへ変換します。
func Recover(logger *slog.Logger, responder *Responder) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.ErrorContext(r.Context(), "panic recovered",
						"panic", recovered,
						"request_id", RequestIDFromContext(r.Context()),
						"stack", string(debug.Stack()),
					)
					responder.Error(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func newRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(value[:])
}
