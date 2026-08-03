CREATE TABLE IF NOT EXISTS urlshortener.users
(
    id             UUID PRIMARY KEY,
    google_sub     VARCHAR(255) NOT NULL,
    email          VARCHAR(320) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    name           TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_google_sub_unique UNIQUE (google_sub)
);
