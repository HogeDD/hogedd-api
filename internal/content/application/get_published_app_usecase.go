package application

import (
	"context"
	"errors"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

// ErrAppNotFound は公開APIから取得可能なAppが存在しないことを表します。
var ErrAppNotFound = errors.New("published app not found")

// AppFinder はAppのslug指定取得に必要なportです。
type AppFinder interface {
	FindBySlug(context.Context, domain.Slug) (*domain.App, bool, error)
}

// GetPublishedAppUseCase は公開済みAppをslugで取得するユースケースです。
type GetPublishedAppUseCase struct {
	finder AppFinder
}

// NewGetPublishedAppUseCase はslug指定取得portを使うユースケースを構築します。
func NewGetPublishedAppUseCase(finder AppFinder) *GetPublishedAppUseCase {
	return &GetPublishedAppUseCase{finder: finder}
}

// Execute はslugを検証し、公開済みAppを返します。
// 不正なslugも存在しないAppと同様に扱い、公開情報の有無を区別しません。
func (uc *GetPublishedAppUseCase) Execute(ctx context.Context, value string) (PublishedAppResult, error) {
	slug, err := domain.NewSlug(value)
	if err != nil {
		return PublishedAppResult{}, ErrAppNotFound
	}

	app, found, err := uc.finder.FindBySlug(ctx, slug)
	if err != nil {
		return PublishedAppResult{}, err
	}
	if !found || !app.PublicationStatus().IsPublic() {
		return PublishedAppResult{}, ErrAppNotFound
	}

	return toPublishedAppResult(app), nil
}
