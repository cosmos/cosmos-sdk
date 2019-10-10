package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics02 "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
)

// Keeper defines each ICS keeper for IBC
type Keeper struct {
	ClientKeeper ics02.Keeper
}

// NewKeeper creates a new ibc Keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		ClientKeeper: ics02.NewKeeper(cdc, key, codespace),
	}
}
