DROP INDEX IF EXISTS urlshortener.idx_links_user_id;

ALTER TABLE urlshortener.links
    DROP COLUMN IF EXISTS user_id;

DROP TABLE IF EXISTS urlshortener.users;
