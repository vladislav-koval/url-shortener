ALTER TABLE analytics.clicks
    ADD COLUMN device_type varchar(100),
    ADD COLUMN os varchar(100),
    ADD COLUMN browser varchar(100),
    ADD COLUMN referer varchar(255);
