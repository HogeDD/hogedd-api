package userhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/iwasawa/hogedd-api/internal/identity"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
)

// ManagementUserGetter は運営境界の認可に必要なUse Case操作です。
type ManagementUserGetter interface {
	Execute(context.Context, identity.Identity) (userapp.ManagementUserResult, error)
}

// ManagementHandler は認可済みの運営Userを返します。
type ManagementHandler struct {
	get       ManagementUserGetter
	responder *httpapi.Responder
}

// NewManagementHandler は運営認可Use CaseをHTTPへ公開するHandlerを構築します。
func NewManagementHandler(get ManagementUserGetter, responder *httpapi.Responder) *ManagementHandler {
	return &ManagementHandler{get: get, responder: responder}
}

// ServeHTTP はGET・HEADによる運営User取得だけを処理します。
func (h *ManagementHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	authenticated, ok := identity.FromContext(r.Context())
	if !ok {
		h.responder.ConcealedNotFound(w, r, "missing_authenticated_identity", nil)
		return
	}
	result, err := h.get.Execute(r.Context(), authenticated)
	if err != nil {
		reason := "management_authorization_failed"
		switch {
		case errors.Is(err, userapp.ErrManagementUserNotFound):
			reason = "management_user_not_found"
		case errors.Is(err, userapp.ErrManagementUserInactive):
			reason = "management_user_inactive"
		case errors.Is(err, userapp.ErrManagementPermissionDenied):
			reason = "management_permission_denied"
		}
		h.responder.ConcealedNotFound(w, r, reason, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		h.responder.PrepareJSON(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	h.responder.JSON(w, http.StatusOK, managementUserResponse{ID: result.ID, Role: result.Role})
}

type managementUserResponse struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}
