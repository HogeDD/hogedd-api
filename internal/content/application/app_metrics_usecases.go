package application

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

var visitorHashPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// ErrInvalidVisitorHash は匿名訪問者hashが所定の形式でないことを表します。
var ErrInvalidVisitorHash = errors.New("invalid visitor hash")

// AppLaunchRecorder は一意なApp起動の記録に必要なportです。
type AppLaunchRecorder interface {
	RecordUniqueLaunch(context.Context, domain.Slug, time.Time, string) (bool, error)
}

// RecordAppLaunchUseCase は匿名訪問者によるApp起動を日単位で記録します。
type RecordAppLaunchUseCase struct {
	recorder AppLaunchRecorder
	now      func() time.Time
}

// NewRecordAppLaunchUseCase は起動記録portとclockからUse Caseを構築します。
func NewRecordAppLaunchUseCase(recorder AppLaunchRecorder, now func() time.Time) *RecordAppLaunchUseCase {
	if now == nil {
		now = time.Now
	}
	return &RecordAppLaunchUseCase{recorder: recorder, now: now}
}

// Execute はslugとvisitor hashを検証して一意な起動を記録します。
func (uc *RecordAppLaunchUseCase) Execute(ctx context.Context, slugValue, visitorHash string) (bool, error) {
	slug, err := domain.NewSlug(slugValue)
	if err != nil {
		return false, ErrAppNotFound
	}
	if !visitorHashPattern.MatchString(visitorHash) {
		return false, ErrInvalidVisitorHash
	}
	return uc.recorder.RecordUniqueLaunch(ctx, slug, uc.now().UTC(), visitorHash)
}

// AppRecommendationReader はランキング候補の取得に必要なportです。
type AppRecommendationReader interface {
	ListLatest(context.Context, int) ([]*domain.App, error)
	ListPopular(context.Context, time.Time, int) ([]*domain.App, error)
	ListTrending(context.Context, time.Time, time.Time, int) ([]*domain.App, error)
	ListRecommended(context.Context) ([]*domain.App, error)
}

// AppRecommendationsResult はApps画面の3種類の推薦枠です。
type AppRecommendationsResult struct {
	Latest   *PublishedAppResult `json:"latest,omitempty"`
	Popular  *PublishedAppResult `json:"popular,omitempty"`
	Trending *PublishedAppResult `json:"trending,omitempty"`
}

// GetAppRecommendationsUseCase は公開Appの推薦枠を重複なく選びます。
type GetAppRecommendationsUseCase struct {
	reader AppRecommendationReader
	now    func() time.Time
}

// NewGetAppRecommendationsUseCase は推薦候補portとclockからUse Caseを構築します。
func NewGetAppRecommendationsUseCase(reader AppRecommendationReader, now func() time.Time) *GetAppRecommendationsUseCase {
	if now == nil {
		now = time.Now
	}
	return &GetAppRecommendationsUseCase{reader: reader, now: now}
}

// Execute は最新・30日人気・2日急上昇を選び、手動推薦で空きを補完します。
func (uc *GetAppRecommendationsUseCase) Execute(ctx context.Context) (AppRecommendationsResult, error) {
	now := uc.now().UTC()
	latest, err := uc.reader.ListLatest(ctx, 10)
	if err != nil {
		return AppRecommendationsResult{}, err
	}
	popular, err := uc.reader.ListPopular(ctx, now.AddDate(0, 0, -30), 10)
	if err != nil {
		return AppRecommendationsResult{}, err
	}
	trending, err := uc.reader.ListTrending(ctx, now.AddDate(0, 0, -2), now.AddDate(0, 0, -30), 10)
	if err != nil {
		return AppRecommendationsResult{}, err
	}
	fallback, err := uc.reader.ListRecommended(ctx)
	if err != nil {
		return AppRecommendationsResult{}, err
	}

	used := map[string]bool{}
	pick := func(groups ...[]*domain.App) *PublishedAppResult {
		for _, group := range groups {
			for _, app := range group {
				if app.PublicationStatus().IsPublic() && !used[app.ID().String()] {
					used[app.ID().String()] = true
					result := toPublishedAppResult(app)
					return &result
				}
			}
		}
		return nil
	}
	result := AppRecommendationsResult{}
	result.Latest = pick(latest, fallback)
	result.Popular = pick(popular, fallback, latest)
	result.Trending = pick(trending, fallback, popular, latest)
	return result, nil
}
