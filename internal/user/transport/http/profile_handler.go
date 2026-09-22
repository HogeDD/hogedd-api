package userhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/iwasawa/hogedd-api/internal/identity"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

const maxProfileBodyBytes = 1024

// CurrentProfileGetter は現在Userのプロフィール取得操作です。
type CurrentProfileGetter interface {
	Execute(context.Context, identity.Identity) (userapp.ProfileResult, error)
}

// CurrentProfileUpdater は現在Userのプロフィール更新操作です。
type CurrentProfileUpdater interface {
	Execute(context.Context, identity.Identity, string) (userapp.ProfileResult, error)
}

// ProfileHandler は現在Userのプロフィールを取得・更新します。
type ProfileHandler struct {
	get       CurrentProfileGetter
	update    CurrentProfileUpdater
	responder *httpapi.Responder
}

// NewProfileHandler はプロフィールHTTP handlerを構築します。
func NewProfileHandler(get CurrentProfileGetter, update CurrentProfileUpdater, responder *httpapi.Responder) *ProfileHandler {
	return &ProfileHandler{get: get, update: update, responder: responder}
}

// ServeHTTP はGET・HEAD・PUTを処理します。
func (h *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		h.getCurrent(w, r)
	case http.MethodPut:
		h.updateCurrent(w, r)
	default:
		w.Header().Set("Allow", "GET, HEAD, PUT")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

func (h *ProfileHandler) getCurrent(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := identity.FromContext(r.Context())
	if !ok {
		h.responder.InternalError(w, r, "load authenticated user context", nil)
		return
	}
	result, err := h.get.Execute(r.Context(), authenticated)
	if h.writeKnownError(w, r, err) {
		return
	}
	h.writeProfile(w, r, http.StatusOK, result)
}

func (h *ProfileHandler) updateCurrent(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := identity.FromContext(r.Context())
	if !ok {
		h.responder.InternalError(w, r, "load authenticated user context", nil)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxProfileBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input struct {
		DisplayName string `json:"display_name"`
	}
	if err := decoder.Decode(&input); err != nil {
		h.responder.Error(w, r, http.StatusBadRequest, "invalid_request", "request body is invalid")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.responder.Error(w, r, http.StatusBadRequest, "invalid_request", "request body is invalid")
		return
	}
	result, err := h.update.Execute(r.Context(), authenticated, input.DisplayName)
	if errors.Is(err, userdomain.ErrDisplayNameRequired) || errors.Is(err, userdomain.ErrDisplayNameTooLong) {
		h.responder.Error(w, r, http.StatusUnprocessableEntity, "invalid_display_name", "display name must be between 1 and 50 characters")
		return
	}
	if h.writeKnownError(w, r, err) {
		return
	}
	h.writeProfile(w, r, http.StatusOK, result)
}

func (h *ProfileHandler) writeKnownError(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, userapp.ErrUserNotFound):
		h.responder.Error(w, r, http.StatusNotFound, "user_not_found", "user not found")
	case errors.Is(err, userapp.ErrProfileNotFound):
		h.responder.Error(w, r, http.StatusNotFound, "profile_not_found", "user profile not found")
	case errors.Is(err, userapp.ErrUserDisabled):
		h.responder.Error(w, r, http.StatusForbidden, "user_disabled", "user is disabled")
	default:
		h.responder.InternalError(w, r, "handle current user profile", err)
	}
	return true
}

func (h *ProfileHandler) writeProfile(w http.ResponseWriter, r *http.Request, status int, result userapp.ProfileResult) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		h.responder.PrepareJSON(w)
		w.WriteHeader(status)
		return
	}
	h.responder.JSON(w, status, struct {
		DisplayName string `json:"display_name"`
	}{DisplayName: result.DisplayName})
}
