CREATE TABLE IF NOT EXISTS urlshortener.users
(
    id             UUID PRIMARY KEY,
    google_sub     VARCHAR(255) NOT NULL,
    email          VARCHAR(320) NOT NULL,
    name           TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_google_sub_unique UNIQUE (google_sub)
);

ALTER TABLE urlshortener.links
    ADD COLUMN user_id UUID
    REFERENCES urlshortener.users(id)
    ON DELETE SET NULL;

CREATE INDEX idx_links_user_id
    ON urlshortener.links(user_id);
