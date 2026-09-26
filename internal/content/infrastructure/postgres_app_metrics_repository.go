package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

// RecordUniqueLaunch は同じ訪問者による同日・同Appの起動を一度だけ集計します。
func (r *PostgresAppRepository) RecordUniqueLaunch(ctx context.Context, slug domain.Slug, metricDate time.Time, visitorHash string) (bool, error) {
	const findPublishedAppIDQuery = `SELECT id FROM content_apps WHERE slug = $1 AND publication_status = 'published'`
	const insertDailyVisitorQuery = `
INSERT INTO content_app_daily_visitors (app_id, metric_date, visitor_hash)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING`
	const incrementDailyLaunchesQuery = `
INSERT INTO content_app_metrics_daily (app_id, metric_date, unique_launches)
VALUES ($1, $2, 1)
ON CONFLICT (app_id, metric_date)
DO UPDATE SET unique_launches = content_app_metrics_daily.unique_launches + 1`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin launch metric transaction: %w", err)
	}
	defer tx.Rollback()
	var appID string
	if err := tx.QueryRowContext(ctx, findPublishedAppIDQuery, slug.String()).Scan(&appID); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("find published app for launch metric: %w", err)
	}
	result, err := tx.ExecContext(ctx, insertDailyVisitorQuery, appID, metricDate.UTC().Format("2006-01-02"), visitorHash)
	if err != nil {
		return false, fmt.Errorf("record unique app visitor: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read unique app visitor result: %w", err)
	}
	if inserted == 0 {
		return false, tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, incrementDailyLaunchesQuery, appID, metricDate.UTC().Format("2006-01-02")); err != nil {
		return false, fmt.Errorf("increment app launch metric: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit launch metric transaction: %w", err)
	}
	return true, nil
}

// ListLatest は公開日時が新しい公開Appを返します。
func (r *PostgresAppRepository) ListLatest(ctx context.Context, limit int) ([]*domain.App, error) {
	const listLatestAppsQuery = `
SELECT a.id, a.slug, a.title, a.description, a.tags, a.publication_status,
       a.published_at, a.development_drive, a.youtube_url
FROM content_apps AS a
WHERE a.publication_status = 'published'
ORDER BY a.published_at DESC, a.slug ASC
LIMIT $1`
	return r.queryRankedApps(ctx, listLatestAppsQuery, limit)
}

// ListPopular は指定日以降のユニーク起動数が多い公開Appを返します。
func (r *PostgresAppRepository) ListPopular(ctx context.Context, since time.Time, limit int) ([]*domain.App, error) {
	const listPopularAppsQuery = `
SELECT a.id, a.slug, a.title, a.description, a.tags, a.publication_status,
       a.published_at, a.development_drive, a.youtube_url
FROM content_apps AS a
JOIN content_app_metrics_daily AS m ON m.app_id = a.id
WHERE a.publication_status = 'published' AND m.metric_date >= $1
GROUP BY a.id
ORDER BY SUM(m.unique_launches) DESC, a.published_at DESC
LIMIT $2`
	return r.queryRankedApps(ctx, listPopularAppsQuery, since.Format("2006-01-02"), limit)
}

// ListTrending は直近2日とそれ以前の期間を比較し、母数3以上の急上昇Appを返します。
func (r *PostgresAppRepository) ListTrending(ctx context.Context, recentSince, baselineSince time.Time, limit int) ([]*domain.App, error) {
	const listTrendingAppsQuery = `
SELECT a.id, a.slug, a.title, a.description, a.tags, a.publication_status,
       a.published_at, a.development_drive, a.youtube_url
FROM content_apps AS a
JOIN content_app_metrics_daily AS m ON m.app_id = a.id
WHERE a.publication_status = 'published' AND m.metric_date >= $1
GROUP BY a.id
HAVING SUM(m.unique_launches) FILTER (WHERE m.metric_date >= $2) >= 3
ORDER BY (
  (SUM(m.unique_launches) FILTER (WHERE m.metric_date >= $2) + 1)::double precision /
  ((COALESCE(SUM(m.unique_launches) FILTER (WHERE m.metric_date < $2), 0)::double precision / 14) + 1)
) * LN((SUM(m.unique_launches) FILTER (WHERE m.metric_date >= $2)) + 1) DESC,
a.published_at DESC
LIMIT $3`
	return r.queryRankedApps(ctx, listTrendingAppsQuery, baselineSince.Format("2006-01-02"), recentSince.Format("2006-01-02"), limit)
}

func (r *PostgresAppRepository) queryRankedApps(ctx context.Context, query string, args ...any) ([]*domain.App, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query ranked apps: %w", err)
	}
	defer rows.Close()
	apps := make([]*domain.App, 0)
	for rows.Next() {
		var idValue, slugValue, title, description, statusValue string
		var tags []string
		var publishedAt sql.NullTime
		var developmentDrive, youtubeURL sql.NullString
		if err := rows.Scan(&idValue, &slugValue, &title, &description, &tags, &statusValue, &publishedAt, &developmentDrive, &youtubeURL); err != nil {
			return nil, fmt.Errorf("scan ranked app: %w", err)
		}
		id, err := domain.RestoreAppID(idValue)
		if err != nil {
			return nil, err
		}
		slug, err := domain.NewSlug(slugValue)
		if err != nil {
			return nil, err
		}
		status, err := domain.ParsePublicationStatus(statusValue)
		if err != nil {
			return nil, err
		}
		app, err := domain.RestoreAppWithID(id, slug, title, description, tags, status, publishedAt.Time, developmentDrive.String, youtubeURL.String)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query ranked apps rows: %w", err)
	}
	return apps, nil
}
