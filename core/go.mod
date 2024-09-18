module cosmossdk.io/core

// Core is meant to have only a dependency on cosmossdk.io/schema, so we can use it as a dependency
// in other modules without having to worry about circular dependencies.

go 1.23

require cosmossdk.io/schema v0.3.0

// Version tagged too early and incompatible with v0.50 (latest at the time of tagging)
retract v0.12.0
