package httpapi

import (
	"context"
	"net/http"

	"github.com/iwasawa/hogedd-api/internal/health"
)

// LivenessService はHTTPヘルスチェックが必要とする最小限のユースケースを表します。
type LivenessService interface {
	Liveness(context.Context) health.Status
}

// HealthHandler はヘルスチェックをHTTPへ公開するアダプターです。
type HealthHandler struct {
	service   LivenessService
	responder *Responder
}

// NewHealthHandler はLivenessServiceと共通Responderを使うHealthHandlerを構築します。
func NewHealthHandler(service LivenessService, responder *Responder) *HealthHandler {
	return &HealthHandler{service: service, responder: responder}
}

// ServeHTTP はGETまたはHEADのヘルスチェックリクエストを処理します。
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	status := h.service.Liveness(r.Context())
	if r.Method == http.MethodHead {
		h.responder.PrepareJSON(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	h.responder.JSON(w, http.StatusOK, struct {
		Status string `json:"status"`
	}{Status: status.Status})
}
