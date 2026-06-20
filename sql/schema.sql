CREATE SCHEMA IF NOT EXISTS reference;

CREATE TABLE IF NOT EXISTS reference.ref_time_slot (
    day_of_week SMALLINT NOT NULL,
    month_of_year SMALLINT NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    period_type VARCHAR NOT NULL,
    pricing_type VARCHAR NOT NULL
);

CREATE VIEW IF NOT EXISTS reference.v_ref_time_slot AS
SELECT
    day_of_week,
    month_of_year,
    period_type,
    pricing_type,
    EXTRACT(HOUR FROM start_time) * 60
      + EXTRACT(MINUTE FROM start_time) AS start_min,
    EXTRACT(HOUR FROM end_time) * 60
      + EXTRACT(MINUTE FROM end_time)
      + 1 AS end_min_excl
FROM reference.ref_time_slot;

