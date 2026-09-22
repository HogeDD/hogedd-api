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

// RegisterHandler は現在の認証主体をHogeDD Userへ登録します。
type RegisterHandler struct {
	register  AuthenticatedUserRegistrar
	responder *httpapi.Responder
}

// NewRegisterHandler はUser登録Use CaseをHTTPへ公開するHandlerを構築します。
func NewRegisterHandler(register AuthenticatedUserRegistrar, responder *httpapi.Responder) *RegisterHandler {
	return &RegisterHandler{register: register, responder: responder}
}

// ServeHTTP はbodyなしのPUTを冪等に処理します。
func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
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
	w.Header().Set("Cache-Control", "no-store")
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
