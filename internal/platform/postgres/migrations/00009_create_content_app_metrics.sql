-- +goose Up
CREATE TABLE content_app_daily_visitors (
    app_id UUID NOT NULL REFERENCES content_apps(id) ON DELETE CASCADE,
    metric_date DATE NOT NULL,
    visitor_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (app_id, metric_date, visitor_hash)
);

CREATE TABLE content_app_metrics_daily (
    app_id UUID NOT NULL REFERENCES content_apps(id) ON DELETE CASCADE,
    metric_date DATE NOT NULL,
    unique_launches BIGINT NOT NULL DEFAULT 0 CHECK (unique_launches >= 0),
    PRIMARY KEY (app_id, metric_date)
);

CREATE INDEX content_app_metrics_daily_date_idx
    ON content_app_metrics_daily (metric_date, app_id);

-- +goose Down
DROP TABLE content_app_metrics_daily;
DROP TABLE content_app_daily_visitors;
