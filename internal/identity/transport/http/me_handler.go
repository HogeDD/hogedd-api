package identityhttp

import (
	"net/http"

	"github.com/iwasawa/hogedd-api/internal/identity"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
)

// MeHandler は現在の認証主体をHTTPレスポンスへ変換します。
type MeHandler struct {
	responder *httpapi.Responder
}

// NewMeHandler は共通Responderを使うMeHandlerを構築します。
func NewMeHandler(responder *httpapi.Responder) *MeHandler {
	return &MeHandler{responder: responder}
}

// ServeHTTP はGETまたはHEADで検証済みIdentityを返します。
func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	authenticated, ok := identity.FromContext(r.Context())
	if !ok {
		h.responder.InternalError(w, r, "load authenticated identity", nil)
		return
	}

	if r.Method == http.MethodHead {
		h.responder.PrepareJSON(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	h.responder.JSON(w, http.StatusOK, struct {
		Issuer  string `json:"issuer"`
		Subject string `json:"subject"`
	}{
		Issuer:  authenticated.Issuer(),
		Subject: authenticated.Subject(),
	})
}
