CREATE TABLE IF NOT EXISTS urlshortener.clicks
(
    id UUID PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    ip_address INET,
    clicked_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_clicks_short_code_clicked_at
    ON urlshortener.clicks (short_code, clicked_at DESC);