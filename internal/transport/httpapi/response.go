package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// Responder はJSONレスポンスと公開エラー形式を一元管理します。
type Responder struct {
	logger *slog.Logger
}

// NewResponder はエンコード失敗を記録するロガーを持つResponderを構築します。
func NewResponder(logger *slog.Logger) *Responder {
	return &Responder{logger: logger}
}

// PrepareJSON はレスポンスをJSONとして返すためのヘッダーを設定します。
func (r *Responder) PrepareJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// JSON はvalueをJSONへエンコードし、指定されたHTTPステータスで返します。
func (r *Responder) JSON(w http.ResponseWriter, status int, value any) {
	r.PrepareJSON(w)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		r.logger.Error("encode HTTP response", "error", err)
	}
}

// Error は安定したエラーコードとrequest IDを持つ共通エラー形式を返します。
func (r *Responder) Error(w http.ResponseWriter, req *http.Request, status int, code, message string) {
	r.JSON(w, status, errorBody{Error: errorDetail{
		Code:      code,
		Message:   message,
		RequestID: RequestIDFromContext(req.Context()),
	}})
}
