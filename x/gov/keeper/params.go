package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GetDepositParams returns the current DepositParams from the global param store
func (keeper Keeper) GetDepositParams(ctx sdk.Context) types.DepositParams {
	var depositParams types.DepositParams
	keeper.paramSpace.Get(ctx, types.ParamStoreKeyDepositParams, &depositParams)
	return depositParams
}

// GetVotingParams returns the current VotingParams from the global param store
func (keeper Keeper) GetVotingParams(ctx sdk.Context) types.VotingParams {
	var votingParams types.VotingParams
	keeper.paramSpace.Get(ctx, types.ParamStoreKeyVotingParams, &votingParams)
	return votingParams
}

// GetTallyParams returns the current TallyParam from the global param store
func (keeper Keeper) GetTallyParams(ctx sdk.Context) types.TallyParams {
	var tallyParams types.TallyParams
	keeper.paramSpace.Get(ctx, types.ParamStoreKeyTallyParams, &tallyParams)
	return tallyParams
}

// SetDepositParams sets DepositParams to the global param store
func (keeper Keeper) SetDepositParams(ctx sdk.Context, depositParams types.DepositParams) {
	keeper.paramSpace.Set(ctx, types.ParamStoreKeyDepositParams, &depositParams)
}

// SetVotingParams sets VotingParams to the global param store
func (keeper Keeper) SetVotingParams(ctx sdk.Context, votingParams types.VotingParams) {
	keeper.paramSpace.Set(ctx, types.ParamStoreKeyVotingParams, &votingParams)
}

// SetTallyParams sets TallyParams to the global param store
func (keeper Keeper) SetTallyParams(ctx sdk.Context, tallyParams types.TallyParams) {
	keeper.paramSpace.Set(ctx, types.ParamStoreKeyTallyParams, &tallyParams)
}
