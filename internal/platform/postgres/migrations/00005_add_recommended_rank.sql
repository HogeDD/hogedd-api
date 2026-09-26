-- +goose Up
ALTER TABLE content_apps ADD COLUMN recommended_rank INTEGER;
CREATE INDEX content_apps_recommended_rank_idx ON content_apps (recommended_rank) WHERE recommended_rank IS NOT NULL;

INSERT INTO content_apps (slug, title, description, tags, publication_status, published_at, development_drive, youtube_url, recommended_rank)
VALUES ('clean-tasks', 'Clean Tasks', 'Clean Architecture の練習として作った、最小構成のタスクアプリ。', ARRAY['Next.js', 'TypeScript', 'Clean Architecture'], 'published', '2026-05-30T00:00:00Z', '学習DD', 'https://youtu.be/5mo0qnPuVTY?si=5comgyazMjAYnOJ9', 1)
ON CONFLICT (slug) DO UPDATE SET recommended_rank = EXCLUDED.recommended_rank;

-- +goose Down
DROP INDEX content_apps_recommended_rank_idx;
ALTER TABLE content_apps DROP COLUMN recommended_rank;
