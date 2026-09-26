package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/iwasawa/hogedd-api/internal/content/application"
	"github.com/iwasawa/hogedd-api/internal/content/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresAppRepository はContent AppをPostgreSQLへ永続化します。
type PostgresAppRepository struct{ db *sql.DB }

// NewPostgresAppRepository はPostgreSQLを使うApp repositoryを構築します。
func NewPostgresAppRepository(db *sql.DB) *PostgresAppRepository {
	return &PostgresAppRepository{db: db}
}

// FindBySlug はSlugに一致するAppを公開判定前の状態で返します。
func (r *PostgresAppRepository) FindBySlug(ctx context.Context, slug domain.Slug) (*domain.App, bool, error) {
	app, _, found, err := r.FindForManagement(ctx, slug)
	return app, found, err
}

// Create は公開準備中Appを作成し、slug重複を業務エラーへ変換します。
func (r *PostgresAppRepository) Create(ctx context.Context, app *domain.App) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO content_apps (slug, title, description, tags, publication_status)
VALUES ($1, $2, $3, $4, $5)`, app.Slug().String(), app.Title(), app.Description(), app.Tags(), string(app.PublicationStatus()))
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return application.ErrAppAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create content app: %w", err)
	}
	return nil
}

// List は管理用に公開状態を問わず全Appを更新日時の降順で返します。
func (r *PostgresAppRepository) List(ctx context.Context) ([]*domain.App, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT slug, title, description, tags, publication_status, published_at, development_drive, youtube_url
FROM content_apps
ORDER BY updated_at DESC, slug ASC`)
	if err != nil {
		return nil, fmt.Errorf("list content apps: %w", err)
	}
	defer rows.Close()
	apps := make([]*domain.App, 0)
	for rows.Next() {
		var slugValue, title, description, statusValue string
		var tags []string
		var publishedAt sql.NullTime
		var developmentDrive, youtubeURL sql.NullString
		if err := rows.Scan(&slugValue, &title, &description, &tags, &statusValue, &publishedAt, &developmentDrive, &youtubeURL); err != nil {
			return nil, fmt.Errorf("scan content app: %w", err)
		}
		slug, err := domain.NewSlug(slugValue)
		if err != nil {
			return nil, fmt.Errorf("restore content app: %w", err)
		}
		status, err := domain.ParsePublicationStatus(statusValue)
		if err != nil {
			return nil, fmt.Errorf("restore content app: %w", err)
		}
		publicationTime := time.Time{}
		if publishedAt.Valid {
			publicationTime = publishedAt.Time
		}
		app, err := domain.RestoreApp(slug, title, description, tags, status, publicationTime, developmentDrive.String, youtubeURL.String)
		if err != nil {
			return nil, fmt.Errorf("restore content app %q: %w", slugValue, err)
		}
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list content apps rows: %w", err)
	}
	return apps, nil
}

// ListRecommended は公開済みでおすすめ順位が設定されたAppを順位順に返します。
func (r *PostgresAppRepository) ListRecommended(ctx context.Context) ([]*domain.App, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT slug, title, description, tags, publication_status, published_at, development_drive, youtube_url FROM content_apps WHERE publication_status = 'published' AND recommended_rank IS NOT NULL ORDER BY recommended_rank ASC, slug ASC`)
	if err != nil {
		return nil, fmt.Errorf("list recommended content apps: %w", err)
	}
	defer rows.Close()
	apps := make([]*domain.App, 0)
	for rows.Next() {
		var slugValue, title, description, statusValue string
		var tags []string
		var publishedAt sql.NullTime
		var developmentDrive, youtubeURL sql.NullString
		if err := rows.Scan(&slugValue, &title, &description, &tags, &statusValue, &publishedAt, &developmentDrive, &youtubeURL); err != nil {
			return nil, fmt.Errorf("scan recommended content app: %w", err)
		}
		slug, err := domain.NewSlug(slugValue)
		if err != nil {
			return nil, err
		}
		status, err := domain.ParsePublicationStatus(statusValue)
		if err != nil {
			return nil, err
		}
		published := time.Time{}
		if publishedAt.Valid {
			published = publishedAt.Time
		}
		app, err := domain.RestoreApp(slug, title, description, tags, status, published, developmentDrive.String, youtubeURL.String)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list recommended content apps rows: %w", err)
	}
	return apps, nil
}

// FindForManagement は管理用に公開状態を問わずAppとversionを取得します。
func (r *PostgresAppRepository) FindForManagement(ctx context.Context, slug domain.Slug) (*domain.App, int64, bool, error) {
	var title, description, statusValue string
	var tags []string
	var publishedAt sql.NullTime
	var developmentDrive, youtubeURL sql.NullString
	var version int64
	err := r.db.QueryRowContext(ctx, `
SELECT title, description, tags, publication_status, published_at, development_drive, youtube_url, version
FROM content_apps WHERE slug = $1`, slug.String()).Scan(&title, &description, &tags, &statusValue, &publishedAt, &developmentDrive, &youtubeURL, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, false, nil
	}
	if err != nil {
		return nil, 0, false, fmt.Errorf("find content app: %w", err)
	}
	status, err := domain.ParsePublicationStatus(statusValue)
	if err != nil {
		return nil, 0, false, fmt.Errorf("restore content app status: %w", err)
	}
	publicationTime := time.Time{}
	if publishedAt.Valid {
		publicationTime = publishedAt.Time
	}
	app, err := domain.RestoreApp(slug, title, description, tags, status, publicationTime, developmentDrive.String, youtubeURL.String)
	if err != nil {
		return nil, 0, false, fmt.Errorf("restore content app %q: %w", slug.String(), err)
	}
	return app, version, true, nil
}

// UpdateDraft はversion一致時だけDraft情報を更新して新しいversionを返します。
func (r *PostgresAppRepository) UpdateDraft(ctx context.Context, app *domain.App, expectedVersion int64) (int64, error) {
	return r.Update(ctx, app, expectedVersion)
}

// Update はversion一致時だけ管理対象Appの全編集項目を保存します。
func (r *PostgresAppRepository) Update(ctx context.Context, app *domain.App, expectedVersion int64) (int64, error) {
	var version int64
	var publishedAt any
	if !app.PublishedAt().IsZero() {
		publishedAt = app.PublishedAt()
	}
	var developmentDrive, youTubeURL any
	if app.DevelopmentDrive() != "" {
		developmentDrive = app.DevelopmentDrive()
	}
	if app.YouTubeURL() != "" {
		youTubeURL = app.YouTubeURL()
	}
	err := r.db.QueryRowContext(ctx, `
UPDATE content_apps
SET title = $1, description = $2, tags = $3, publication_status = $4,
    published_at = $5, development_drive = $6, youtube_url = $7,
    version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE slug = $8 AND version = $9
RETURNING version`, app.Title(), app.Description(), app.Tags(), string(app.PublicationStatus()), publishedAt, developmentDrive, youTubeURL, app.Slug().String(), expectedVersion).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, application.ErrAppVersionConflict
	}
	if err != nil {
		return 0, fmt.Errorf("update draft content app: %w", err)
	}
	return version, nil
}

// Publish はversion一致時だけAppを公開済みに更新して新しいversionを返します。
func (r *PostgresAppRepository) Publish(ctx context.Context, app *domain.App, expectedVersion int64) (int64, error) {
	var version int64
	err := r.db.QueryRowContext(ctx, `
UPDATE content_apps
SET publication_status = $1, published_at = $2, development_drive = $3, youtube_url = $4,
    version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE slug = $5 AND publication_status IN ('preparing', 'private') AND version = $6
RETURNING version`, string(app.PublicationStatus()), app.PublishedAt(), app.DevelopmentDrive(), app.YouTubeURL(), app.Slug().String(), expectedVersion).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, application.ErrAppVersionConflict
	}
	if err != nil {
		return 0, fmt.Errorf("publish content app: %w", err)
	}
	return version, nil
}

// MakePrivate はversion一致時だけAppを非公開に更新して新しいversionを返します。
func (r *PostgresAppRepository) MakePrivate(ctx context.Context, app *domain.App, expectedVersion int64) (int64, error) {
	var version int64
	err := r.db.QueryRowContext(ctx, `
UPDATE content_apps
SET publication_status = 'private', version = version + 1, updated_at = CURRENT_TIMESTAMP
WHERE slug = $1 AND publication_status = 'published' AND version = $2
RETURNING version`, app.Slug().String(), expectedVersion).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, application.ErrAppVersionConflict
	}
	if err != nil {
		return 0, fmt.Errorf("make content app private: %w", err)
	}
	return version, nil
}
