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
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
)

const maxManagementAppBodyBytes = 16 << 10

// ManagementAppsAuthorizer は運営User認可に必要な操作です。
type ManagementAppsAuthorizer interface {
	Execute(context.Context, identity.Identity) (userapp.ManagementUserResult, error)
}

// ManagementAppsLister は運営App一覧に必要な操作です。
type ManagementAppsLister interface {
	Execute(context.Context) ([]contentapp.ManagementAppResult, error)
}

// PreparingAppCreator は公開準備中App作成に必要な操作です。
type PreparingAppCreator interface {
	Execute(context.Context, string, string, string, []string) (contentapp.ManagementAppResult, error)
}

// ManagementAppsHandler は認可済み運営UserによるApp一覧・作成を処理します。
type ManagementAppsHandler struct {
	authorize ManagementAppsAuthorizer
	list      ManagementAppsLister
	create    PreparingAppCreator
	responder *httpapi.Responder
}

// NewManagementAppsHandler は運営認可とContent Use CaseをHTTPへ接続します。
func NewManagementAppsHandler(authorize ManagementAppsAuthorizer, list ManagementAppsLister, create PreparingAppCreator, responder *httpapi.Responder) *ManagementAppsHandler {
	return &ManagementAppsHandler{authorize: authorize, list: list, create: create, responder: responder}
}

// ServeHTTP はGETによる一覧とPOSTによるdraft作成を処理します。
func (h *ManagementAppsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		h.listApps(w, r)
	case http.MethodPost:
		h.createApp(w, r)
	default:
		w.Header().Set("Allow", "GET, HEAD, POST")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

func (h *ManagementAppsHandler) authorized(w http.ResponseWriter, r *http.Request) bool {
	authenticated, ok := identity.FromContext(r.Context())
	if !ok {
		h.responder.ConcealedNotFound(w, r, "missing_authenticated_identity", nil)
		return false
	}
	if _, err := h.authorize.Execute(r.Context(), authenticated); err != nil {
		h.responder.ConcealedNotFound(w, r, "management_authorization_failed", err)
		return false
	}
	return true
}

func (h *ManagementAppsHandler) listApps(w http.ResponseWriter, r *http.Request) {
	apps, err := h.list.Execute(r.Context())
	if err != nil {
		h.responder.InternalError(w, r, "list management apps", err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		h.responder.PrepareJSON(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	h.responder.JSON(w, http.StatusOK, struct {
		Data []contentapp.ManagementAppResult `json:"data"`
	}{Data: apps})
}

func (h *ManagementAppsHandler) createApp(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxManagementAppBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input struct {
		Slug        string   `json:"slug"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := decoder.Decode(&input); err != nil {
		h.responder.Error(w, r, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.responder.Error(w, r, http.StatusBadRequest, "invalid_request", "request body must contain one JSON object")
		return
	}
	app, err := h.create.Execute(r.Context(), input.Slug, input.Title, input.Description, input.Tags)
	switch {
	case errors.Is(err, contentapp.ErrAppAlreadyExists):
		h.responder.Error(w, r, http.StatusConflict, "app_already_exists", "app already exists")
	case errors.Is(err, domain.ErrInvalidSlug), errors.Is(err, domain.ErrTitleRequired), errors.Is(err, domain.ErrDescriptionRequired):
		h.responder.Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "app input is invalid")
	case errors.Is(err, domain.ErrTitleTooLong), errors.Is(err, domain.ErrDescriptionTooLong), errors.Is(err, domain.ErrInvalidTags):
		h.responder.Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "app input is invalid")
	case err != nil:
		h.responder.InternalError(w, r, "create management app", err)
	default:
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Location", "/v1/management/apps/"+app.Slug)
		h.responder.JSON(w, http.StatusCreated, app)
	}
}
