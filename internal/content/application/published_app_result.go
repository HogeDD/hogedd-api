package application

import (
	"time"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

// PublishedAppResult は公開APIへ渡せるAppの取得結果です。
// Domain ModelをTransportへ直接渡さず、Application層で公開可能な情報へ写像します。
type PublishedAppResult struct {
	Slug             string
	Title            string
	Description      string
	PublishedAt      time.Time
	Tags             []string
	DevelopmentDrive string
	YouTubeURL       string
}

func toPublishedAppResult(app *domain.App) PublishedAppResult {
	return PublishedAppResult{
		Slug:             app.Slug().String(),
		Title:            app.Title(),
		Description:      app.Description(),
		PublishedAt:      app.PublishedAt(),
		Tags:             app.Tags(),
		DevelopmentDrive: app.DevelopmentDrive(),
		YouTubeURL:       app.YouTubeURL(),
	}
}
