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
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
)

type managementAuthorizerStub struct{ err error }

func (s managementAuthorizerStub) Execute(context.Context, identity.Identity) (userapp.ManagementUserResult, error) {
	return userapp.ManagementUserResult{ID: "user-1", Role: "owner"}, s.err
}

type managementAppsStub struct {
	apps    []contentapp.ManagementAppResult
	created contentapp.ManagementAppResult
	err     error
}

func (s *managementAppsStub) Execute(context.Context) ([]contentapp.ManagementAppResult, error) {
	return s.apps, s.err
}

type preparingAppCreatorStub struct{ *managementAppsStub }

func (s preparingAppCreatorStub) Execute(context.Context, string, string, string, []string) (contentapp.ManagementAppResult, error) {
	return s.created, s.err
}

func managementAppsRequest(t *testing.T, method, body string, authorizer managementAuthorizerStub, apps *managementAppsStub) *httptest.ResponseRecorder {
	t.Helper()
	responder := httpapi.NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := NewManagementAppsHandler(authorizer, apps, preparingAppCreatorStub{apps}, responder)
	authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
	request := httptest.NewRequest(method, "/v1/management/apps", strings.NewReader(body))
	request = request.WithContext(identity.NewContext(request.Context(), authenticated))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestManagementAppsHandlerListsAndCreates(t *testing.T) {
	apps := &managementAppsStub{apps: []contentapp.ManagementAppResult{{Slug: "draft", Status: "preparing"}}, created: contentapp.ManagementAppResult{Slug: "new-app", Status: "preparing"}}
	list := managementAppsRequest(t, http.MethodGet, "", managementAuthorizerStub{}, apps)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"slug":"draft"`) {
		t.Fatalf("GET = %d %s", list.Code, list.Body.String())
	}
	create := managementAppsRequest(t, http.MethodPost, `{"slug":"new-app","title":"New","description":"Description","tags":[]}`, managementAuthorizerStub{}, apps)
	if create.Code != http.StatusCreated || create.Header().Get("Location") != "/v1/management/apps/new-app" {
		t.Fatalf("POST = %d %s", create.Code, create.Body.String())
	}
}

func TestManagementAppsHandlerConcealsUnauthorizedUser(t *testing.T) {
	recorder := managementAppsRequest(t, http.MethodGet, "", managementAuthorizerStub{err: userapp.ErrManagementPermissionDenied}, &managementAppsStub{})
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestManagementAppsHandlerRejectsInvalidBody(t *testing.T) {
	recorder := managementAppsRequest(t, http.MethodPost, `{"slug":`, managementAuthorizerStub{}, &managementAppsStub{})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status/body = %d %s", recorder.Code, recorder.Body.String())
	}
}
