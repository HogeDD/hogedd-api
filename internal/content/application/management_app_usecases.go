package application

import (
	"context"
	"errors"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

// ErrAppAlreadyExists は同じslugのAppが既に存在することを表します。
var ErrAppAlreadyExists = errors.New("app already exists")

// AppCreator はApp作成に必要な永続化portです。
type AppCreator interface {
	Create(context.Context, *domain.App) error
}

// ManagementAppResult は運営画面へ返す公開前情報を含むAppです。
type ManagementAppResult struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Tags        []string `json:"tags"`
}

// CreatePreparingAppUseCase は公開準備中Appを作成します。
type CreatePreparingAppUseCase struct{ creator AppCreator }

// NewCreatePreparingAppUseCase は永続化portを使う作成Use Caseを構築します。
func NewCreatePreparingAppUseCase(creator AppCreator) *CreatePreparingAppUseCase {
	return &CreatePreparingAppUseCase{creator: creator}
}

// Execute は入力をdomain modelで検証し、公開準備中Appとして保存します。
func (uc *CreatePreparingAppUseCase) Execute(ctx context.Context, slugValue, title, description string, tags []string) (ManagementAppResult, error) {
	slug, err := domain.NewSlug(slugValue)
	if err != nil {
		return ManagementAppResult{}, err
	}
	app, err := domain.NewPreparingApp(slug, title, description, tags)
	if err != nil {
		return ManagementAppResult{}, err
	}
	if err := uc.creator.Create(ctx, app); err != nil {
		return ManagementAppResult{}, err
	}
	return toManagementAppResult(app), nil
}

// ListManagementAppsUseCase は公開状態を問わずApp一覧を取得します。
type ListManagementAppsUseCase struct{ lister AppLister }

// NewListManagementAppsUseCase は一覧取得portを使う運営一覧Use Caseを構築します。
func NewListManagementAppsUseCase(lister AppLister) *ListManagementAppsUseCase {
	return &ListManagementAppsUseCase{lister: lister}
}

// Execute はdraftを含む全Appを運営用結果へ変換します。
func (uc *ListManagementAppsUseCase) Execute(ctx context.Context) ([]ManagementAppResult, error) {
	apps, err := uc.lister.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ManagementAppResult, 0, len(apps))
	for _, app := range apps {
		result = append(result, toManagementAppResult(app))
	}
	return result, nil
}

func toManagementAppResult(app *domain.App) ManagementAppResult {
	return ManagementAppResult{Slug: app.Slug().String(), Title: app.Title(), Description: app.Description(), Tags: app.Tags(), Status: string(app.PublicationStatus())}
}
