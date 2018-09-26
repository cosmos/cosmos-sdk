package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params Params
}

// HubDefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState, sdata types.GenesisState) {
	for _, validator := range sdata.Validators {
		keeper.addPubkey(ctx, validator.GetConsPubKey())
	}

	keeper.paramstore.SetStruct(ctx, &data.Params)
}
