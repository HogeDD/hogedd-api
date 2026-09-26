package application

import (
	"context"
	"errors"
	"time"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

// ErrAppAlreadyExists は同じslugのAppが既に存在することを表します。
var ErrAppAlreadyExists = errors.New("app already exists")

// ErrManagementAppNotFound は管理対象Appが存在しないことを表します。
var ErrManagementAppNotFound = errors.New("management app not found")

// ErrAppVersionConflict は取得後にAppが別更新されていることを表します。
var ErrAppVersionConflict = errors.New("app version conflict")

// AppCreator はApp作成に必要な永続化portです。
type AppCreator interface {
	Create(context.Context, *domain.App) error
}

// ManagementAppResult は運営画面へ返す公開前情報を含むAppです。
type ManagementAppResult struct {
	Slug             string   `json:"slug"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Status           string   `json:"status"`
	Tags             []string `json:"tags"`
	Version          int64    `json:"version,omitempty"`
	PublishedAt      string   `json:"published_at,omitempty"`
	DevelopmentDrive string   `json:"development_drive,omitempty"`
	YouTubeURL       string   `json:"youtube_url,omitempty"`
}

// ManagementAppPublisher はApp公開の楽観的ロック更新に必要なportです。
type ManagementAppPublisher interface {
	ManagementAppEditor
	Publish(context.Context, *domain.App, int64) (int64, error)
}

// PublishManagementAppUseCase は公開準備中Appを公開します。
type PublishManagementAppUseCase struct {
	publisher ManagementAppPublisher
	now       func() time.Time
}

// NewPublishManagementAppUseCase は公開永続化portとsystem clockからUse Caseを構築します。
func NewPublishManagementAppUseCase(publisher ManagementAppPublisher, now func() time.Time) *PublishManagementAppUseCase {
	if now == nil {
		now = time.Now
	}
	return &PublishManagementAppUseCase{publisher: publisher, now: now}
}

// Execute は公開済みなら現在値を返し、Draftなら公開条件を検証して保存します。
func (uc *PublishManagementAppUseCase) Execute(ctx context.Context, value, developmentDrive, youTubeURL string, expectedVersion int64) (ManagementAppResult, error) {
	if expectedVersion < 1 {
		return ManagementAppResult{}, ErrAppVersionConflict
	}
	slug, err := domain.NewSlug(value)
	if err != nil {
		return ManagementAppResult{}, ErrManagementAppNotFound
	}
	app, version, found, err := uc.publisher.FindForManagement(ctx, slug)
	if err != nil {
		return ManagementAppResult{}, err
	}
	if !found {
		return ManagementAppResult{}, ErrManagementAppNotFound
	}
	if app.PublicationStatus().IsPublic() {
		result := toManagementAppResult(app)
		result.Version = version
		return result, nil
	}
	if version != expectedVersion {
		return ManagementAppResult{}, ErrAppVersionConflict
	}
	if err := app.Publish(uc.now().UTC(), developmentDrive, youTubeURL); err != nil {
		return ManagementAppResult{}, err
	}
	version, err = uc.publisher.Publish(ctx, app, expectedVersion)
	if err != nil {
		return ManagementAppResult{}, err
	}
	result := toManagementAppResult(app)
	result.Version = version
	return result, nil
}

// ManagementAppEditor は管理用Appの取得と楽観的ロック更新に必要なportです。
type ManagementAppEditor interface {
	FindForManagement(context.Context, domain.Slug) (*domain.App, int64, bool, error)
	UpdateDraft(context.Context, *domain.App, int64) (int64, error)
}

// GetManagementAppUseCase は管理用App詳細を取得します。
type GetManagementAppUseCase struct{ editor ManagementAppEditor }

// NewGetManagementAppUseCase は管理用取得portからUse Caseを構築します。
func NewGetManagementAppUseCase(editor ManagementAppEditor) *GetManagementAppUseCase {
	return &GetManagementAppUseCase{editor: editor}
}

// Execute はSlugを検証し、存在するAppと現在versionを返します。
func (uc *GetManagementAppUseCase) Execute(ctx context.Context, value string) (ManagementAppResult, error) {
	slug, err := domain.NewSlug(value)
	if err != nil {
		return ManagementAppResult{}, ErrManagementAppNotFound
	}
	app, version, found, err := uc.editor.FindForManagement(ctx, slug)
	if err != nil {
		return ManagementAppResult{}, err
	}
	if !found {
		return ManagementAppResult{}, ErrManagementAppNotFound
	}
	result := toManagementAppResult(app)
	result.Version = version
	return result, nil
}

// UpdateManagementAppUseCase は公開準備中Appの表示情報を更新します。
type UpdateManagementAppUseCase struct{ editor ManagementAppEditor }

// NewUpdateManagementAppUseCase は管理用更新portからUse Caseを構築します。
func NewUpdateManagementAppUseCase(editor ManagementAppEditor) *UpdateManagementAppUseCase {
	return &UpdateManagementAppUseCase{editor: editor}
}

// Execute は現在versionを照合し、Draft情報を更新します。
func (uc *UpdateManagementAppUseCase) Execute(ctx context.Context, value, title, description string, tags []string, expectedVersion int64) (ManagementAppResult, error) {
	if expectedVersion < 1 {
		return ManagementAppResult{}, ErrAppVersionConflict
	}
	slug, err := domain.NewSlug(value)
	if err != nil {
		return ManagementAppResult{}, ErrManagementAppNotFound
	}
	app, _, found, err := uc.editor.FindForManagement(ctx, slug)
	if err != nil {
		return ManagementAppResult{}, err
	}
	if !found {
		return ManagementAppResult{}, ErrManagementAppNotFound
	}
	if err := app.UpdateDraftDetails(title, description, tags); err != nil {
		return ManagementAppResult{}, err
	}
	version, err := uc.editor.UpdateDraft(ctx, app, expectedVersion)
	if err != nil {
		return ManagementAppResult{}, err
	}
	result := toManagementAppResult(app)
	result.Version = version
	return result, nil
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
	result := ManagementAppResult{Slug: app.Slug().String(), Title: app.Title(), Description: app.Description(), Tags: app.Tags(), Status: string(app.PublicationStatus())}
	if app.PublicationStatus().IsPublic() {
		result.PublishedAt = app.PublishedAt().UTC().Format(time.RFC3339)
		result.DevelopmentDrive = app.DevelopmentDrive()
		result.YouTubeURL = app.YouTubeURL()
	}
	return result
}
