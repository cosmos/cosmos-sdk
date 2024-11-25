module cosmossdk.io/indexer/postgres

// NOTE: we are staying on an earlier version of golang to avoid problems building
// with older codebases.
go 1.12

// NOTE: cosmossdk.io/schema should be the only dependency here
// so there are no problems building this with any version of the SDK.
// This module should only use the golang standard library (database/sql)
// and cosmossdk.io/indexer/base.
require cosmossdk.io/schema v0.3.0

replace cosmossdk.io/schema => ../../schema
