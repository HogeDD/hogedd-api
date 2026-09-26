package contenthttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	contentapp "github.com/iwasawa/hogedd-api/internal/content/application"
	"github.com/iwasawa/hogedd-api/internal/content/domain"
	"github.com/iwasawa/hogedd-api/internal/identity"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
)

// ManagementAppPublisher は管理用App公開に必要な操作です。
type ManagementAppPublisher interface {
	Execute(context.Context, string, string, string, int64) (contentapp.ManagementAppResult, error)
}

// ManagementAppPublicationHandler はAppの公開操作を処理します。
type ManagementAppPublicationHandler struct {
	authorize ManagementAppsAuthorizer
	publish   ManagementAppPublisher
	responder *httpapi.Responder
}

// NewManagementAppPublicationHandler は運営認可と公開Use CaseをHTTPへ接続します。
func NewManagementAppPublicationHandler(authorize ManagementAppsAuthorizer, publish ManagementAppPublisher, responder *httpapi.Responder) *ManagementAppPublicationHandler {
	return &ManagementAppPublicationHandler{authorize: authorize, publish: publish, responder: responder}
}

// ServeHTTP はPUTによる冪等な公開操作を処理します。
func (h *ManagementAppPublicationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := identity.FromContext(r.Context())
	if !ok {
		h.responder.ConcealedNotFound(w, r, "missing_authenticated_identity", nil)
		return
	}
	if _, err := h.authorize.Execute(r.Context(), authenticated); err != nil {
		h.responder.ConcealedNotFound(w, r, "management_authorization_failed", err)
		return
	}
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", "PUT")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxManagementAppBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input struct {
		DevelopmentDrive string `json:"development_drive"`
		YouTubeURL       string `json:"youtube_url"`
		Version          int64  `json:"version"`
	}
	if err := decoder.Decode(&input); err != nil {
		h.responder.Error(w, r, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.responder.Error(w, r, http.StatusBadRequest, "invalid_request", "request body must contain one JSON object")
		return
	}
	result, err := h.publish.Execute(r.Context(), r.PathValue("slug"), input.DevelopmentDrive, input.YouTubeURL, input.Version)
	switch {
	case errors.Is(err, contentapp.ErrManagementAppNotFound):
		h.responder.ConcealedNotFound(w, r, "management_app_not_found", err)
	case errors.Is(err, contentapp.ErrAppVersionConflict):
		h.responder.Error(w, r, http.StatusConflict, "app_version_conflict", "app was updated by another request")
	case errors.Is(err, domain.ErrPublishedAtRequired), errors.Is(err, domain.ErrDevelopmentDriveRequired), errors.Is(err, domain.ErrInvalidYouTubeURL):
		h.responder.Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "publication input is invalid")
	case err != nil:
		h.responder.InternalError(w, r, "publish management app", err)
	default:
		w.Header().Set("Cache-Control", "no-store")
		h.responder.JSON(w, http.StatusOK, result)
	}
}
