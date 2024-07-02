CREATE SCHEMA IF NOT EXISTS indexer;

CREATE TABLE IF NOT EXISTS indexer.indexed_modules
(
    module_name TEXT NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS block
(
    block_number BIGINT NOT NULL PRIMARY KEY,
    header       JSONB  NULL
);

CREATE TABLE IF NOT EXISTS tx
(
    id           BIGSERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL REFERENCES block (block_number),
    tx_index     BIGINT NOT NULL,
    data         JSONB  NOT NULL
);

CREATE TABLE IF NOT EXISTS event
(
    id           BIGSERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL REFERENCES block (block_number),
    tx_index     BIGINT NOT NULL REFERENCES tx (tx_index),
    msg_idx      BIGINT NOT NULL,
    event_idx    BIGINT NOT NULL,
    type         TEXT   NOT NULL,
    data         JSONB  NOT NULL
);


