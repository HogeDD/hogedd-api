package userhttp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iwasawa/hogedd-api/internal/identity"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
)

type registerUseCaseStub struct {
	result userapp.RegisteredUserResult
	err    error
	token  string
}

type getUserUseCaseStub struct {
	result userapp.RegisteredUserResult
	err    error
}

func (s getUserUseCaseStub) Execute(context.Context, identity.Identity) (userapp.RegisteredUserResult, error) {
	return s.result, s.err
}

func (s *registerUseCaseStub) Execute(
	_ context.Context,
	_ identity.Identity,
	token string,
) (userapp.RegisteredUserResult, error) {
	s.token = token
	return s.result, s.err
}

type verifierStub struct {
	identity identity.Identity
}

func (s verifierStub) Verify(context.Context, string) (identity.Identity, error) {
	return s.identity, nil
}

func authenticatedHandler(t *testing.T, useCase *registerUseCaseStub) http.Handler {
	t.Helper()
	authenticated, err := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	if err != nil {
		t.Fatal(err)
	}
	responder := httpapi.NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := NewMeHandler(getUserUseCaseStub{}, useCase, responder)
	return httpapi.AuthenticateBearer(verifierStub{identity: authenticated}, responder)(handler)
}

func TestMeHandlerGetsCurrentUser(t *testing.T) {
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	responder := httpapi.NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := NewMeHandler(getUserUseCaseStub{result: userapp.RegisteredUserResult{
		ID: "0199-user", Email: "owner@example.com", EmailVerified: true,
		Role: "member", Status: "active", CreatedAt: now, UpdatedAt: now,
	}}, &registerUseCaseStub{}, responder)
	protected := httpapi.AuthenticateBearer(verifierStub{identity: authenticated}, responder)(handler)
	request := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	protected.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "owner@example.com") {
		t.Fatalf("status/body = %d, %s", recorder.Code, recorder.Body.String())
	}
}

func TestMeHandlerReturnsNotFound(t *testing.T) {
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|missing")
	responder := httpapi.NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := NewMeHandler(getUserUseCaseStub{err: userapp.ErrUserNotFound}, &registerUseCaseStub{}, responder)
	protected := httpapi.AuthenticateBearer(verifierStub{identity: authenticated}, responder)(handler)
	request := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	protected.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestMeHandlerCreatesUser(t *testing.T) {
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	useCase := &registerUseCaseStub{result: userapp.RegisteredUserResult{
		ID:            "0199-user",
		Email:         "owner@example.com",
		EmailVerified: true,
		Role:          "member",
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
		Created:       true,
	}}
	handler := authenticatedHandler(t, useCase)
	request := httptest.NewRequest(http.MethodPut, "/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Location") != "/v1/users/me" {
		t.Errorf("Location = %q", recorder.Header().Get("Location"))
	}
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q", recorder.Header().Get("Cache-Control"))
	}
	if useCase.token != "access-token" {
		t.Errorf("Access Token = %q", useCase.token)
	}
	if strings.Contains(recorder.Body.String(), "access-token") {
		t.Error("response contains Access Token")
	}
}

func TestMeHandlerReturnsOKForExistingUser(t *testing.T) {
	useCase := &registerUseCaseStub{result: userapp.RegisteredUserResult{Created: false}}
	handler := authenticatedHandler(t, useCase)
	request := httptest.NewRequest(http.MethodPut, "/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestMeHandlerRejectsMethodAndBody(t *testing.T) {
	for _, test := range []struct {
		name   string
		method string
		body   string
		status int
	}{
		{name: "method", method: http.MethodPost, status: http.StatusMethodNotAllowed},
		{name: "body", method: http.MethodPut, body: `{}`, status: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := authenticatedHandler(t, &registerUseCaseStub{})
			request := httptest.NewRequest(test.method, "/v1/users/me", strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer access-token")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestMeHandlerMapsProfileFailureToBadGateway(t *testing.T) {
	useCase := &registerUseCaseStub{err: errors.New("wrapped: " + userapp.ErrProfileUnavailable.Error())}
	useCase.err = errors.Join(useCase.err, userapp.ErrProfileUnavailable)
	handler := authenticatedHandler(t, useCase)
	request := httptest.NewRequest(http.MethodPut, "/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}
