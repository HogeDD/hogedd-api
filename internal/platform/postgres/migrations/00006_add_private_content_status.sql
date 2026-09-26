-- +goose Up
ALTER TABLE content_apps DROP CONSTRAINT content_apps_status_check;
ALTER TABLE content_apps ADD CONSTRAINT content_apps_status_check CHECK (publication_status IN ('preparing', 'published', 'private'));
ALTER TABLE content_apps DROP CONSTRAINT content_apps_publication_fields_check;
ALTER TABLE content_apps ADD CONSTRAINT content_apps_publication_fields_check CHECK (
    (publication_status = 'preparing' AND published_at IS NULL AND development_drive IS NULL AND youtube_url IS NULL)
    OR
    (publication_status IN ('published', 'private') AND published_at IS NOT NULL AND development_drive IS NOT NULL AND youtube_url IS NOT NULL)
);

-- +goose Down
UPDATE content_apps SET publication_status = 'preparing', published_at = NULL, development_drive = NULL, youtube_url = NULL WHERE publication_status = 'private';
ALTER TABLE content_apps DROP CONSTRAINT content_apps_publication_fields_check;
ALTER TABLE content_apps ADD CONSTRAINT content_apps_publication_fields_check CHECK (
    (publication_status = 'preparing' AND published_at IS NULL AND development_drive IS NULL AND youtube_url IS NULL)
    OR
    (publication_status = 'published' AND published_at IS NOT NULL AND development_drive IS NOT NULL AND youtube_url IS NOT NULL)
);
ALTER TABLE content_apps DROP CONSTRAINT content_apps_status_check;
ALTER TABLE content_apps ADD CONSTRAINT content_apps_status_check CHECK (publication_status IN ('preparing', 'published'));
