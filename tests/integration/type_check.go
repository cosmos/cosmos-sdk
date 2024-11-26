package integration

import (
	db "github.com/cosmos/cosmos-db"

	coretesting "cosmossdk.io/core/testing"
)

// This file contains a list of type checks that are used to ensure that implementations
// matches the interface. We do not do those type checks directly in the components to
// avoid to bring in more dependencies than needed.
var (
	_ db.DB = (*coretesting.MemDB)(nil)
)
