package postgres

// baseSQL is the base SQL that is always included in the schema.
const baseSQL = `
CREATE OR REPLACE FUNCTION nanos_to_timestamptz(nanos bigint) RETURNS timestamptz AS $$
    SELECT to_timestamp(nanos / 1000000000) + (nanos / 1000000000) * INTERVAL '1 microsecond'
$$ LANGUAGE SQL IMMUTABLE;

CREATE TABLE IF NOT EXISTS block
(
    number BIGINT NOT NULL PRIMARY KEY,
    header JSONB  NULL
);

CREATE TABLE IF NOT EXISTS tx
(
    id             BIGSERIAL PRIMARY KEY,
    block_number   BIGINT NOT NULL REFERENCES block (number),
    index_in_block BIGINT NOT NULL,
    data           JSONB NULL,
    bytes          BYTEA NULL 
);

CREATE TABLE IF NOT EXISTS event
(
    id           BIGSERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL REFERENCES block (number),
    block_stage  INTEGER NOT NULL,
    tx_index     BIGINT NOT NULL,
    msg_index    BIGINT NOT NULL,
    event_index  BIGINT NOT NULL,
    type         TEXT   NULL,
    data         JSONB  NULL
);
`
