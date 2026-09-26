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

type managementAppPublisherStub struct {
	result contentapp.ManagementAppResult
	err    error
}

func (s managementAppPublisherStub) Execute(context.Context, string, string, string, int64) (contentapp.ManagementAppResult, error) {
	return s.result, s.err
}

func TestManagementAppPublicationHandlerPublishesAndMapsConflict(t *testing.T) {
	for _, test := range []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "published", wantStatus: http.StatusOK},
		{name: "conflict", err: contentapp.ErrAppVersionConflict, wantStatus: http.StatusConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			responder := httpapi.NewResponder(slog.New(slog.NewTextHandler(io.Discard, nil)))
			publisher := managementAppPublisherStub{result: contentapp.ManagementAppResult{Slug: "draft", Status: "published", Version: 2}, err: test.err}
			handler := NewManagementAppPublicationHandler(managementAuthorizerStub{}, publisher, responder)
			authenticated, _ := identity.New("https://hogedd.jp.auth0.com/", "auth0|owner")
			request := httptest.NewRequest(http.MethodPut, "/v1/management/apps/draft/publication", strings.NewReader(`{"development_drive":"学習DD","youtube_url":"https://youtu.be/video","version":1}`))
			request.SetPathValue("slug", "draft")
			request = request.WithContext(identity.NewContext(request.Context(), authenticated))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status/body = %d %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
