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
