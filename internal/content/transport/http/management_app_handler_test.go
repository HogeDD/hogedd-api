package contenthttp

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contentapp "github.com/iwasawa/hogedd-api/internal/content/application"
	"github.com/iwasawa/hogedd-api/internal/identity"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
)

type managementAppGetterStub struct {
	result contentapp.ManagementAppResult
	err    error
}

func (s managementAppGetterStub) Execute(context.Context, string) (contentapp.ManagementAppResult, error) {
	return s.result, s.err
}

type managementAppUpdaterStub struct{ managementAppGetterStub }

func (s managementAppUpdaterStub) Execute(context.Context, string, string, string, []string, int64) (contentapp.ManagementAppResult, error) {
	return s.result, s.err
}

func managementAppRequest(t *testing.T, method, body string, authorizer managementAuthorizerStub, result contentapp.ManagementAppResult, useCaseError error) *httptest.ResponseRecorder {
	t.Helper()
	responder := httpapi.NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	useCase := managementAppGetterStub{result: result, err: useCaseError}
	handler := NewManagementAppHandler(authorizer, useCase, managementAppUpdaterStub{useCase}, responder)
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	request := httptest.NewRequest(method, "/v1/management/apps/draft", strings.NewReader(body))
	request.SetPathValue("slug", "draft")
	request = request.WithContext(identity.NewContext(request.Context(), authenticated))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestManagementAppHandlerGetsAndUpdatesDraft(t *testing.T) {
	result := contentapp.ManagementAppResult{Slug: "draft", Title: "Draft", Description: "Description", Status: "preparing", Tags: []string{}, Version: 2}
	get := managementAppRequest(t, http.MethodGet, "", managementAuthorizerStub{}, result, nil)
	if get.Code != http.StatusOK || !strings.Contains(get.Body.String(), `"version":2`) {
		t.Fatalf("GET = %d %s", get.Code, get.Body.String())
	}
	update := managementAppRequest(t, http.MethodPut, `{"title":"Draft","description":"Description","tags":[],"version":1}`, managementAuthorizerStub{}, result, nil)
	if update.Code != http.StatusOK {
		t.Fatalf("PUT = %d %s", update.Code, update.Body.String())
	}
}

func TestManagementAppHandlerMapsConflictAndConcealsAuthorization(t *testing.T) {
	conflict := managementAppRequest(t, http.MethodPut, `{"title":"Draft","description":"Description","tags":[],"version":1}`, managementAuthorizerStub{}, contentapp.ManagementAppResult{}, contentapp.ErrAppVersionConflict)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status = %d", conflict.Code)
	}
	concealed := managementAppRequest(t, http.MethodGet, "", managementAuthorizerStub{err: context.Canceled}, contentapp.ManagementAppResult{}, nil)
	if concealed.Code != http.StatusNotFound {
		t.Fatalf("concealed status = %d", concealed.Code)
	}
}
