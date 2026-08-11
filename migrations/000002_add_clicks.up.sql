CREATE SCHEMA IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.clicks
(
    id UUID PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    ip_address INET,
    clicked_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_clicks_short_code_clicked_at
    ON analytics.clicks (short_code, clicked_at DESC);