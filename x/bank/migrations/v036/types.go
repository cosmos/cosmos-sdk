// Package v036 is used for legacy migration scripts. Actual migration scripts
// for v036 have been removed, but the v039->v042 migration script still
// references types from this file, so we're keeping it for now.
// DONTCOVER
package v036

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ModuleName = "supply"

type (
	GenesisState struct {
		Supply sdk.Coins `json:"supply" yaml:"supply"`
	}
)
