package postgres

// BaseSQL is the base SQL that is always included in the schema.
const BaseSQL = `
CREATE OR REPLACE FUNCTION nanos_to_timestamptz(nanos bigint) RETURNS timestamptz AS $$
    SELECT to_timestamp(nanos / 1000000000) + (nanos / 1000000000) * INTERVAL '1 microsecond'
$$ LANGUAGE SQL IMMUTABLE;
`
