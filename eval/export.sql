-- Freeze the evaluation stream: one pulse per line, in stream (modified) order.
-- Usage: psql "$DSN" -At -f eval/export.sql > eval/data/pulses.jsonl
SELECT json_build_object(
    'id', p.id,
    'indicators', COALESCE(
        (SELECT json_agg(json_build_object('indicator', i.indicator, 'type', i.type) ORDER BY i.id)
         FROM indicators i WHERE i.pulse_id = p.id),
        '[]'::json)
)
FROM pulses p
ORDER BY p.modified, p.id;
