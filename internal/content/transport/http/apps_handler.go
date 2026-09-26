package contenthttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/iwasawa/hogedd-api/internal/content/application"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
)

// PublishedAppsLister は公開アプリ一覧のUse Caseに必要な操作です。
type PublishedAppsLister interface {
	Execute(context.Context) ([]application.PublishedAppResult, error)
}

// PublishedAppGetter は公開アプリの1件取得Use Caseに必要な操作です。
type PublishedAppGetter interface {
	Execute(context.Context, string) (application.PublishedAppResult, error)
}

// AppsHandler は公開アプリの取得結果をHTTPレスポンスへ変換します。
type AppsHandler struct {
	list        PublishedAppsLister
	recommended PublishedAppsLister
	get         PublishedAppGetter
	responder   *httpapi.Responder
}

// NewAppsHandler は一覧・1件取得のUse Caseを持つHTTP Handlerを構築します。
func NewAppsHandler(list PublishedAppsLister, get PublishedAppGetter, recommended PublishedAppsLister, responder *httpapi.Responder) *AppsHandler {
	return &AppsHandler{list: list, get: get, recommended: recommended, responder: responder}
}

// Recommended はおすすめ公開アプリ一覧へのGETまたはHEADを処理します。
func (h *AppsHandler) Recommended(w http.ResponseWriter, r *http.Request) {
	if !h.allowRead(w, r) {
		return
	}
	apps, err := h.recommended.Execute(r.Context())
	if err != nil {
		h.responder.InternalError(w, r, "load recommended apps", err)
		return
	}
	data := make([]publishedAppResponse, 0, len(apps))
	for _, app := range apps {
		data = append(data, toPublishedAppResponse(app))
	}
	h.writeJSON(w, r, http.StatusOK, struct {
		Data []publishedAppResponse `json:"data"`
	}{Data: data})
}

// List は公開アプリ一覧へのGETまたはHEADリクエストを処理します。
func (h *AppsHandler) List(w http.ResponseWriter, r *http.Request) {
	if !h.allowRead(w, r) {
		return
	}

	apps, err := h.list.Execute(r.Context())
	if err != nil {
		h.responder.InternalError(w, r, "load published apps", err)
		return
	}

	data := make([]publishedAppResponse, 0, len(apps))
	for _, app := range apps {
		data = append(data, toPublishedAppResponse(app))
	}
	h.writeJSON(w, r, http.StatusOK, struct {
		Data []publishedAppResponse `json:"data"`
	}{Data: data})
}

// Get はpath valueのslugを使って公開アプリを1件取得します。
func (h *AppsHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !h.allowRead(w, r) {
		return
	}

	app, err := h.get.Execute(r.Context(), r.PathValue("slug"))
	if errors.Is(err, application.ErrAppNotFound) {
		h.responder.Error(w, r, http.StatusNotFound, "app_not_found", "app not found")
		return
	}
	if err != nil {
		h.responder.InternalError(w, r, "load published app", err)
		return
	}
	h.writeJSON(w, r, http.StatusOK, toPublishedAppResponse(app))
}

func (h *AppsHandler) allowRead(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	w.Header().Set("Allow", "GET, HEAD")
	h.responder.Error(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	return false
}

func (h *AppsHandler) writeJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	if r.Method == http.MethodHead {
		h.responder.PrepareJSON(w)
		w.WriteHeader(status)
		return
	}
	h.responder.JSON(w, status, value)
}

type publishedAppResponse struct {
	Slug             string   `json:"slug"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Status           string   `json:"status"`
	PublishedAt      string   `json:"published_at"`
	Tags             []string `json:"tags"`
	DevelopmentDrive string   `json:"development_drive"`
	YouTubeURL       string   `json:"youtube_url"`
}

func toPublishedAppResponse(app application.PublishedAppResult) publishedAppResponse {
	return publishedAppResponse{
		Slug:             app.Slug,
		Title:            app.Title,
		Description:      app.Description,
		Status:           "published",
		PublishedAt:      app.PublishedAt.UTC().Format(time.RFC3339),
		Tags:             app.Tags,
		DevelopmentDrive: app.DevelopmentDrive,
		YouTubeURL:       app.YouTubeURL,
	}
}
