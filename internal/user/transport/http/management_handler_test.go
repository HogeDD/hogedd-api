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

	"github.com/iwasawa/hogedd-api/internal/identity"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
)

type managementUserGetterStub struct {
	result userapp.ManagementUserResult
	err    error
}

func (s managementUserGetterStub) Execute(context.Context, identity.Identity) (userapp.ManagementUserResult, error) {
	return s.result, s.err
}

func managementHTTPHandler(t *testing.T, getter managementUserGetterStub) http.Handler {
	t.Helper()
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	responder := httpapi.NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	return httpapi.ConcealBearer(verifierStub{identity: authenticated}, responder)(NewManagementHandler(getter, responder))
}

func TestManagementHandlerReturnsAuthorizedUser(t *testing.T) {
	handler := managementHTTPHandler(t, managementUserGetterStub{result: userapp.ManagementUserResult{ID: "user-1", Role: "owner"}})
	request := httptest.NewRequest(http.MethodGet, "/v1/management/me", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"role":"owner"`) {
		t.Fatalf("status/body = %d, %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q", recorder.Header().Get("Cache-Control"))
	}
}

func TestManagementHandlerConcealsAuthorizationFailures(t *testing.T) {
	errorsToConceal := []error{
		userapp.ErrManagementUserNotFound,
		userapp.ErrManagementUserInactive,
		userapp.ErrManagementPermissionDenied,
		errors.New("database unavailable"),
	}
	for _, concealed := range errorsToConceal {
		handler := managementHTTPHandler(t, managementUserGetterStub{err: concealed})
		request := httptest.NewRequest(http.MethodGet, "/v1/management/me", nil)
		request.Header.Set("Authorization", "Bearer access-token")
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNotFound || !strings.Contains(recorder.Body.String(), `"code":"not_found"`) {
			t.Errorf("error/status/body = %v, %d, %s", concealed, recorder.Code, recorder.Body.String())
		}
	}
}

func TestManagementHandlerSupportsHeadAndRejectsMethodForAuthorizedUser(t *testing.T) {
	handler := managementHTTPHandler(t, managementUserGetterStub{result: userapp.ManagementUserResult{ID: "user-1", Role: "admin"}})
	for _, tt := range []struct {
		method     string
		wantStatus int
	}{
		{method: http.MethodHead, wantStatus: http.StatusOK},
		{method: http.MethodPost, wantStatus: http.StatusMethodNotAllowed},
	} {
		request := httptest.NewRequest(tt.method, "/v1/management/me", nil)
		request.Header.Set("Authorization", "Bearer access-token")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != tt.wantStatus {
			t.Errorf("%s status = %d, want %d", tt.method, recorder.Code, tt.wantStatus)
		}
		if tt.method == http.MethodHead && recorder.Body.Len() != 0 {
			t.Errorf("HEAD body = %q", recorder.Body.String())
		}
	}
}
