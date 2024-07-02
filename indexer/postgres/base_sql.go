package postgres

const baseSql = `
CREATE OR REPLACE FUNCTION nanos_to_timestamptz(nanos bigint) RETURNS timestamptz AS $$
    SELECT to_timestamp(nanos / 1000000000) + nanos * INTERVAL '1 microsecond'
$$ LANGUAGE SQL IMMUTABLE;
`
