WITH params AS (
    SELECT ?::TIMESTAMP AS ts
),

jours AS (
    SELECT
        CAST(ts AS DATE) + CAST(i AS INTEGER) AS jour,
        ts
    FROM params, range(0, 15) AS r(i)
),

plages AS (
    SELECT
        j.ts,
        j.jour,
        r.period_type,
        r.pricing_type,
        CAST(j.jour AS TIMESTAMP)
          + r.start_min * INTERVAL '1 minute' AS start_slot,
        CAST(j.jour AS TIMESTAMP)
          + r.end_min_excl * INTERVAL '1 minute' AS end_slot
    FROM jours j
    JOIN reference.v_ref_time_slot r
      ON r.day_of_week = CAST(strftime(j.jour, '%w') AS INTEGER)
     AND r.month_of_year = EXTRACT(MONTH FROM j.jour)
),

plages_ordonnees AS (
    SELECT
        *,
        LAG(pricing_type) OVER (ORDER BY start_slot, end_slot) AS prev_pricing_type,
        LAG(period_type) OVER (ORDER BY start_slot, end_slot) AS prev_period_type,
        LAG(end_slot) OVER (ORDER BY start_slot, end_slot) AS prev_end_slot
    FROM plages
),

plages_marquees AS (
    SELECT
        *,
        CASE
            WHEN prev_pricing_type = pricing_type
             AND prev_period_type = period_type
             AND prev_end_slot = start_slot
            THEN 0
            ELSE 1
        END AS new_group
    FROM plages_ordonnees
),

plages_groupees AS (
    SELECT
        *,
        SUM(new_group) OVER (ORDER BY start_slot, end_slot ROWS UNBOUNDED PRECEDING) AS group_id
    FROM plages_marquees
),

plages_fusionnees AS (
    SELECT
        ts,
        pricing_type,
        period_type,
        MIN(start_slot) AS start_slot,
        MAX(end_slot) AS end_slot
    FROM plages_groupees
    GROUP BY ts, group_id, pricing_type, period_type
),

current_slot AS (
    SELECT *
    FROM plages_fusionnees
    WHERE ts >= start_slot
      AND ts <  end_slot
    ORDER BY start_slot
    LIMIT 1
)

SELECT
    p.pricing_type,
    p.period_type,
    p.start_slot,
    p.end_slot,
    date_diff('minute', p.start_slot, p.end_slot) AS duration_minutes
FROM plages_fusionnees p
JOIN current_slot c ON TRUE
WHERE p.start_slot >= c.end_slot
ORDER BY p.start_slot
LIMIT ?;
