package contenthttp

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/iwasawa/hogedd-api/internal/content/application"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
)

// AppLaunchRecorder は匿名App起動の記録に必要な操作です。
type AppLaunchRecorder interface {
	Execute(context.Context, string, string) (bool, error)
}

// AppRecommendationsGetter はApp推薦枠の取得に必要な操作です。
type AppRecommendationsGetter interface {
	Execute(context.Context) (application.AppRecommendationsResult, error)
}

// AppMetricsHandler は匿名起動記録と推薦枠取得をHTTPへ接続します。
type AppMetricsHandler struct {
	record      AppLaunchRecorder
	recommend   AppRecommendationsGetter
	ingestToken string
	responder   *httpapi.Responder
}

// NewAppMetricsHandler は計測Use Case、推薦Use Case、計測用秘密値からHandlerを構築します。
func NewAppMetricsHandler(record AppLaunchRecorder, recommend AppRecommendationsGetter, ingestToken string, responder *httpapi.Responder) *AppMetricsHandler {
	return &AppMetricsHandler{record: record, recommend: recommend, ingestToken: ingestToken, responder: responder}
}

// RecordLaunch はBFFから送られた匿名起動を記録します。
func (h *AppMetricsHandler) RecordLaunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	provided := r.Header.Get("X-HogeDD-Metrics-Token")
	if h.ingestToken == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(h.ingestToken)) != 1 {
		h.responder.Error(w, r, http.StatusNotFound, "not_found", "not found")
		return
	}
	var input struct {
		VisitorHash string `json:"visitor_hash"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		h.responder.Error(w, r, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	_, err := h.record.Execute(r.Context(), r.PathValue("slug"), input.VisitorHash)
	switch {
	case errors.Is(err, application.ErrAppNotFound):
		h.responder.Error(w, r, http.StatusNotFound, "app_not_found", "app not found")
	case errors.Is(err, application.ErrInvalidVisitorHash):
		h.responder.Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "visitor hash is invalid")
	case err != nil:
		h.responder.InternalError(w, r, "record app launch", err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// Recommendations は最新・人気・急上昇の推薦枠を返します。
func (h *AppMetricsHandler) Recommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	result, err := h.recommend.Execute(r.Context())
	if err != nil {
		h.responder.InternalError(w, r, "load app recommendations", err)
		return
	}
	type response struct {
		Latest   *publishedAppResponse `json:"latest,omitempty"`
		Popular  *publishedAppResponse `json:"popular,omitempty"`
		Trending *publishedAppResponse `json:"trending,omitempty"`
	}
	value := response{}
	if result.Latest != nil {
		item := toPublishedAppResponse(*result.Latest)
		value.Latest = &item
	}
	if result.Popular != nil {
		item := toPublishedAppResponse(*result.Popular)
		value.Popular = &item
	}
	if result.Trending != nil {
		item := toPublishedAppResponse(*result.Trending)
		value.Trending = &item
	}
	w.Header().Set("Cache-Control", "public, s-maxage=300, stale-while-revalidate=600")
	if r.Method == http.MethodHead {
		h.responder.PrepareJSON(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	h.responder.JSON(w, http.StatusOK, value)
}
