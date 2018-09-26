package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// InitGenesis initializes the keeper's address to pubkey map.
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	for _, validator := range data.Validators {
		keeper.addPubkey(ctx, validator.GetConsPubKey())
	}
	return
}
