CREATE TABLE pulses (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    author_name TEXT,
    modified TEXT,
    created TEXT,
    revision INTEGER,
    tlp TEXT,
    public INTEGER,
    adversary TEXT
);

CREATE TABLE indicators (
    id BIGINT PRIMARY KEY,
    pulse_id TEXT NOT NULL REFERENCES pulses (id) ON DELETE CASCADE,
    indicator TEXT NOT NULL,
    type TEXT,
    created TEXT,
    title TEXT,
    is_active INTEGER
);

CREATE TABLE edges (
    id BIGSERIAL PRIMARY KEY,
    source_pulse_id TEXT NOT NULL REFERENCES pulses (id) ON DELETE CASCADE,
    target_pulse_id TEXT NOT NULL REFERENCES pulses (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_pulse_id, target_pulse_id)
);
