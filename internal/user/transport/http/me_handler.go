package userhttp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/iwasawa/hogedd-api/internal/identity"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
)

// AuthenticatedUserRegistrar は認証済みUser登録に必要なUse Case操作です。
type AuthenticatedUserRegistrar interface {
	Execute(context.Context, identity.Identity, string) (userapp.RegisteredUserResult, error)
}

// CurrentUserGetter は現在の認証主体に紐づくUser取得に必要なUse Case操作です。
type CurrentUserGetter interface {
	Execute(context.Context, identity.Identity) (userapp.RegisteredUserResult, error)
}

// MeHandler は現在の認証主体に紐づくHogeDD Userを取得・登録します。
type MeHandler struct {
	get       CurrentUserGetter
	register  AuthenticatedUserRegistrar
	responder *httpapi.Responder
}

// NewMeHandler はUser取得・登録Use CaseをHTTPへ公開するHandlerを構築します。
func NewMeHandler(
	get CurrentUserGetter,
	register AuthenticatedUserRegistrar,
	responder *httpapi.Responder,
) *MeHandler {
	return &MeHandler{get: get, register: register, responder: responder}
}

// ServeHTTP はGET・HEADによる取得とbodyなしPUTによる冪等登録を処理します。
func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		h.getCurrent(w, r)
	case http.MethodPut:
		h.registerCurrent(w, r)
	default:
		w.Header().Set("Allow", "GET, HEAD, PUT")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

func (h *MeHandler) getCurrent(w http.ResponseWriter, r *http.Request) {
	authenticated, ok := identity.FromContext(r.Context())
	if !ok {
		h.responder.InternalError(w, r, "load authenticated user context", nil)
		return
	}
	result, err := h.get.Execute(r.Context(), authenticated)
	if errors.Is(err, userapp.ErrUserNotFound) {
		h.responder.Error(w, r, http.StatusNotFound, "user_not_found", "user not found")
		return
	}
	if err != nil {
		h.responder.InternalError(w, r, "get current user", err)
		return
	}
	h.writeUser(w, r, http.StatusOK, result)
}

func (h *MeHandler) registerCurrent(w http.ResponseWriter, r *http.Request) {
	if hasRequestBody(r) {
		h.responder.Error(w, r, http.StatusBadRequest, "request_body_not_allowed", "request body is not allowed")
		return
	}

	authenticated, identityOK := identity.FromContext(r.Context())
	accessToken, tokenOK := httpapi.AccessTokenFromContext(r.Context())
	if !identityOK || !tokenOK {
		h.responder.InternalError(w, r, "load authenticated registration context", nil)
		return
	}

	result, err := h.register.Execute(r.Context(), authenticated, accessToken)
	if errors.Is(err, userapp.ErrProfileUnavailable) || errors.Is(err, userapp.ErrProfileIdentityMismatch) {
		h.responder.Error(w, r, http.StatusBadGateway, "identity_provider_unavailable", "identity provider temporarily unavailable")
		return
	}
	if err != nil {
		h.responder.InternalError(w, r, "register authenticated user", err)
		return
	}

	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
		w.Header().Set("Location", "/v1/users/me")
	}
	h.writeUser(w, r, status, result)
}

func (h *MeHandler) writeUser(w http.ResponseWriter, r *http.Request, status int, result userapp.RegisteredUserResult) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		h.responder.PrepareJSON(w)
		w.WriteHeader(status)
		return
	}
	h.responder.JSON(w, status, registeredUserResponse{
		ID:            result.ID,
		Email:         result.Email,
		EmailVerified: result.EmailVerified,
		Role:          result.Role,
		Status:        result.Status,
		CreatedAt:     result.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     result.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

func hasRequestBody(r *http.Request) bool {
	if r.Body == nil {
		return false
	}
	var buffer [1]byte
	read, _ := io.ReadFull(r.Body, buffer[:])
	return read > 0
}

type registeredUserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
