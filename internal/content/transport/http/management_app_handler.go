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

// ManagementAppGetter は管理用App詳細取得に必要な操作です。
type ManagementAppGetter interface {
	Execute(context.Context, string) (contentapp.ManagementAppResult, error)
}

// ManagementAppUpdater は管理用Draft更新に必要な操作です。
type ManagementAppUpdater interface {
	Execute(context.Context, string, string, string, []string, string, string, string, int64) (contentapp.ManagementAppResult, error)
}

// ManagementAppHandler は単一Appの管理用取得・更新を処理します。
type ManagementAppHandler struct {
	authorize ManagementAppsAuthorizer
	get       ManagementAppGetter
	update    ManagementAppUpdater
	responder *httpapi.Responder
}

// NewManagementAppHandler は運営認可と詳細取得・更新Use CaseをHTTPへ接続します。
func NewManagementAppHandler(authorize ManagementAppsAuthorizer, get ManagementAppGetter, update ManagementAppUpdater, responder *httpapi.Responder) *ManagementAppHandler {
	return &ManagementAppHandler{authorize: authorize, get: get, update: update, responder: responder}
}

// ServeHTTP はGET／HEADによる取得とPUTによるDraft更新を処理します。
func (h *ManagementAppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := identity.FromContext(r.Context())
	if !ok {
		h.responder.ConcealedNotFound(w, r, "missing_authenticated_identity", nil)
		return
	}
	if _, err := h.authorize.Execute(r.Context(), authenticated); err != nil {
		h.responder.ConcealedNotFound(w, r, "management_authorization_failed", err)
		return
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		h.getApp(w, r)
	case http.MethodPut:
		h.updateApp(w, r)
	default:
		w.Header().Set("Allow", "GET, HEAD, PUT")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

func (h *ManagementAppHandler) getApp(w http.ResponseWriter, r *http.Request) {
	result, err := h.get.Execute(r.Context(), r.PathValue("slug"))
	if errors.Is(err, contentapp.ErrManagementAppNotFound) {
		h.responder.ConcealedNotFound(w, r, "management_app_not_found", err)
		return
	}
	if err != nil {
		h.responder.InternalError(w, r, "get management app", err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		h.responder.PrepareJSON(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	h.responder.JSON(w, http.StatusOK, result)
}

func (h *ManagementAppHandler) updateApp(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxManagementAppBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input struct {
		Title            string   `json:"title"`
		Description      string   `json:"description"`
		Tags             []string `json:"tags"`
		Status           string   `json:"status"`
		DevelopmentDrive string   `json:"development_drive"`
		YouTubeURL       string   `json:"youtube_url"`
		Version          int64    `json:"version"`
	}
	if err := decoder.Decode(&input); err != nil {
		h.responder.Error(w, r, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.responder.Error(w, r, http.StatusBadRequest, "invalid_request", "request body must contain one JSON object")
		return
	}
	result, err := h.update.Execute(r.Context(), r.PathValue("slug"), input.Title, input.Description, input.Tags, input.Status, input.DevelopmentDrive, input.YouTubeURL, input.Version)
	switch {
	case errors.Is(err, contentapp.ErrManagementAppNotFound):
		h.responder.ConcealedNotFound(w, r, "management_app_not_found", err)
	case errors.Is(err, contentapp.ErrAppVersionConflict):
		h.responder.Error(w, r, http.StatusConflict, "app_version_conflict", "app was updated by another request")
	case errors.Is(err, domain.ErrTitleRequired), errors.Is(err, domain.ErrDescriptionRequired), errors.Is(err, domain.ErrTitleTooLong), errors.Is(err, domain.ErrDescriptionTooLong), errors.Is(err, domain.ErrInvalidTags), errors.Is(err, domain.ErrDevelopmentDriveRequired), errors.Is(err, domain.ErrInvalidYouTubeURL), errors.Is(err, domain.ErrPublishedAtRequired):
		h.responder.Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "app input is invalid")
	case err != nil:
		h.responder.InternalError(w, r, "update management app", err)
	default:
		w.Header().Set("Cache-Control", "no-store")
		h.responder.JSON(w, http.StatusOK, result)
	}
}
