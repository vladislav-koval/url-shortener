ALTER TABLE analytics.clicks
    ADD COLUMN country varchar(100),
    ADD COLUMN city varchar(100),
    DROP COLUMN ip_address;