package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
)

// Keeper defines each ICS keeper for IBC
type Keeper struct {
	ClientKeeper client.Keeper
}

// NewKeeper creates a new ibc Keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		ClientKeeper: client.NewKeeper(cdc, key, codespace),
	}
}
