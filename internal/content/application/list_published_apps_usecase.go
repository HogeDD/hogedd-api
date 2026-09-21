package application

import (
	"context"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

// AppLister はAppの一覧取得に必要なportです。
type AppLister interface {
	List(context.Context) ([]*domain.App, error)
}

// ListPublishedAppsUseCase は公開済みAppの一覧を取得するユースケースです。
type ListPublishedAppsUseCase struct {
	lister AppLister
}

// NewListPublishedAppsUseCase は一覧取得portを使うユースケースを構築します。
func NewListPublishedAppsUseCase(lister AppLister) *ListPublishedAppsUseCase {
	return &ListPublishedAppsUseCase{lister: lister}
}

// Execute は公開済みAppを公開用の取得結果へ変換して返します。
func (uc *ListPublishedAppsUseCase) Execute(ctx context.Context) ([]PublishedAppResult, error) {
	apps, err := uc.lister.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]PublishedAppResult, 0, len(apps))
	for _, app := range apps {
		if !app.PublicationStatus().IsPublic() {
			continue
		}
		result = append(result, toPublishedAppResult(app))
	}

	return result, nil
}
