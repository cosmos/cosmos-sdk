// Package exported defines the `generate()` and `authenticate()` functions for
// capability keys as defined in https://github.com/cosmos/ics/tree/master/spec/ics-005-port-allocation#data-structures.
package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Generate creates a new object-capability key, which must
// be returned by the outer-layer function.
type Generate func() sdk.CapabilityKey

// Authenticate defines an authentication function defined by
// each module to authenticate their own.
type Authenticate func(key sdk.CapabilityKey) bool
