package integration

import (
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// CreateMultiStore is a helper for setting up multiple stores for provided modules.
// Deprecated: use github.com/cosmos/cosmos-sdk/types/module/testutil.CreateMultiStore instead.
var CreateMultiStore = moduletestutil.CreateMultiStore
