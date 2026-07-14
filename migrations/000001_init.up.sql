CREATE SCHEMA IF NOT EXISTS urlshortener;

CREATE TABLE IF NOT EXISTS urlshortener.links
(
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT links_short_code_unique UNIQUE (short_code)
);

