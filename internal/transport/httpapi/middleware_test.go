package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareStackAppliesCommonAndEndpointMiddlewareInOrder(t *testing.T) {
	t.Parallel()

	var calls []string
	record := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, name+":before")
				next.ServeHTTP(w, r)
				calls = append(calls, name+":after")
			})
		}
	}
	stack := NewMiddlewareStack(record("common"))
	handler := stack.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls = append(calls, "handler")
		w.WriteHeader(http.StatusNoContent)
	}), record("endpoint"))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"common:before", "endpoint:before", "handler", "endpoint:after", "common:after"}
	if len(calls) != len(want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("calls = %v, want %v", calls, want)
		}
	}
}

func TestRecover(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	responder := NewResponder(logger)
	panicHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	})
	handler := Chain(panicHandler, RequestID(), Recover(logger, responder))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var body errorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "internal_error" {
		t.Errorf("error code = %q", body.Error.Code)
	}
	if body.Error.RequestID == "" {
		t.Error("request ID is empty")
	}
}

func TestRequestIDRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	handler := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), RequestID())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "invalid request id")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got == "" || got == "invalid request id" {
		t.Fatalf("generated X-Request-ID = %q", got)
	}
}
