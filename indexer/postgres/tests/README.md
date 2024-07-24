# PostgreSQL Indexer Tests

The majority of tests for the PostgreSQL indexer are stored in this separate `tests` go module to keep the main indexer module free of dependencies on any particular PostgreSQL driver. This allows users to choose their own driver and integrate the indexer free of any dependency conflict concerns.