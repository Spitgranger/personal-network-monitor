CREATE TABLE IF NOT EXISTS event (
    schema_version INTEGER NOT NULL,
    event_id TEXT NOT NULL PRIMARY KEY,
    device_id TEXT NOT NULL,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    check_id INTEGER NOT NULL,
    kind TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('success', 'failure')),
    duration_ms INTEGER NULL CHECK (duration_ms is NULL OR duration_ms >= 0),
    error_code TEXT NULL,
    value INTEGER NULL,
    unit TEXT NULL,
    observed_at TIMESTAMP NOT NULL,

    CONSTRAINT unique_id_seq UNIQUE (device_id, sequence)
);

CREATE INDEX idx_event_observed_on ON event(observed_at);

CREATE TABLE IF NOT EXISTS agent_state (
    agent_id TEXT NOT NULL PRIMARY KEY,
    next_sequence INTEGER NOT NULL CHECK (next_sequence > 0)
);
