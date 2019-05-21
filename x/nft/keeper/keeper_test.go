package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	k   Keeper
}
