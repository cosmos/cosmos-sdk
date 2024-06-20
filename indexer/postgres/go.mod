module cosmossdk.io/indexer/postgres

// NOTE: we are staying on an earlier version of golang to avoid problems building
// with older codebases.
go 1.12

require (
	// NOTE: cosmossdk.io/schema and cosmossdk.io/indexer/base should be the only dependencies here
	// so there are no problems building this with any version of the SDK.
	// This module should only use the golang standard library (database/sql)
	// and cosmossdk.io/indexer/base.
	cosmossdk.io/schema v0.0.0
	cosmossdk.io/indexer/base v0.0.0-00010101000000-000000000000
)

replace cosmossdk.io/indexer/base => ../base
replace cosmossdk.io/schema => ../../schema
