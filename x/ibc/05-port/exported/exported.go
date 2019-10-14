// Package exported defines the `generate()` and `authenticate()` functions for
// capability keys as defined in https://github.com/cosmos/ics/tree/master/spec/ics-005-port-allocation#data-structures.
package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Generate
type Generate func() sdk.CapabilityKey

// Authenticate
type Authenticate func(key sdk.CapabilityKey) bool
