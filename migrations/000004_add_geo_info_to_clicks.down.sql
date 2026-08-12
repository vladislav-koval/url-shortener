ALTER TABLE analytics.clicks
    DROP COLUMN country,
    DROP COLUMN city,
    ADD COLUMN ip_address INET