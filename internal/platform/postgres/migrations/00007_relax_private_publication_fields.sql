-- +goose Up
ALTER TABLE content_apps DROP CONSTRAINT content_apps_publication_fields_check;
ALTER TABLE content_apps ADD CONSTRAINT content_apps_publication_fields_check CHECK (
    publication_status <> 'published'
    OR (published_at IS NOT NULL AND development_drive IS NOT NULL AND youtube_url IS NOT NULL)
);

-- +goose Down
UPDATE content_apps
SET development_drive = COALESCE(development_drive, ''), youtube_url = COALESCE(youtube_url, '')
WHERE publication_status = 'private';
ALTER TABLE content_apps DROP CONSTRAINT content_apps_publication_fields_check;
ALTER TABLE content_apps ADD CONSTRAINT content_apps_publication_fields_check CHECK (
    (publication_status = 'preparing' AND published_at IS NULL AND development_drive IS NULL AND youtube_url IS NULL)
    OR
    (publication_status IN ('published', 'private') AND published_at IS NOT NULL AND development_drive IS NOT NULL AND youtube_url IS NOT NULL)
);
