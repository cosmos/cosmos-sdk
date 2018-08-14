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
func HubDefaultGenesisState() GenesisState {
	return GenesisState{
		Params: HubDefaultParams(),
	}
}

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState, sdata types.GenesisState) error {
	for _, validator := range sdata.Validators {
		keeper.addPubkey(ctx, validator.GetPubKey())
	}

	p := data.Params
	keeper.paramstore.Set(ctx, maxEvidenceAgeKey, p.MaxEvidenceAge)
	keeper.paramstore.Set(ctx, signedBlocksWindowKey, p.SignedBlocksWindow)
	keeper.paramstore.Set(ctx, minSignedPerWindowKey, p.MinSignedPerWindow)
	keeper.paramstore.Set(ctx, doubleSignUnbondDurationKey, p.DoubleSignUnbondDuration)
	keeper.paramstore.Set(ctx, downtimeUnbondDurationKey, p.DowntimeUnbondDuration)
	keeper.paramstore.Set(ctx, slashFractionDoubleSignKey, p.SlashFractionDoubleSign)
	keeper.paramstore.Set(ctx, slashFractionDowntimeKey, p.SlashFractionDowntime)
	return nil
}
