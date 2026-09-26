-- +goose Up
CREATE TABLE content_apps (
    slug TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    publication_status TEXT NOT NULL DEFAULT 'preparing',
    published_at TIMESTAMPTZ,
    development_drive TEXT,
    youtube_url TEXT,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT content_apps_slug_not_blank CHECK (BTRIM(slug) <> ''),
    CONSTRAINT content_apps_title_not_blank CHECK (BTRIM(title) <> ''),
    CONSTRAINT content_apps_description_not_blank CHECK (BTRIM(description) <> ''),
    CONSTRAINT content_apps_status_check CHECK (publication_status IN ('preparing', 'published')),
    CONSTRAINT content_apps_version_positive CHECK (version > 0),
    CONSTRAINT content_apps_publication_fields_check CHECK (
        (publication_status = 'preparing' AND published_at IS NULL AND development_drive IS NULL AND youtube_url IS NULL)
        OR
        (publication_status = 'published' AND published_at IS NOT NULL AND development_drive IS NOT NULL AND youtube_url IS NOT NULL)
    )
);

-- +goose Down
DROP TABLE content_apps;
