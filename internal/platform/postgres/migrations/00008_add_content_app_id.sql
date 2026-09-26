-- +goose Up
ALTER TABLE content_apps ADD COLUMN id UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE content_apps DROP CONSTRAINT content_apps_pkey;
ALTER TABLE content_apps ADD CONSTRAINT content_apps_pkey PRIMARY KEY (id);
ALTER TABLE content_apps ADD CONSTRAINT content_apps_slug_key UNIQUE (slug);

-- +goose Down
ALTER TABLE content_apps DROP CONSTRAINT content_apps_slug_key;
ALTER TABLE content_apps DROP CONSTRAINT content_apps_pkey;
ALTER TABLE content_apps ADD CONSTRAINT content_apps_pkey PRIMARY KEY (slug);
ALTER TABLE content_apps DROP COLUMN id;
