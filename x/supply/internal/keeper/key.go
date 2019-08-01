package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// DefaultCodespace from the supply module
var DefaultCodespace sdk.CodespaceType = types.ModuleName

// Keys for supply store
// Items are stored with the following key: values
//
// - 0x00: Supply
var (
	SupplyKey = []byte{0x00}
)
