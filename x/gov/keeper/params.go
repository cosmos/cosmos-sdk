package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// GetDepositParams returns the current DepositParams from the global param store
func (keeper Keeper) GetDepositParams(ctx sdk.Context) v1.DepositParams {
	var depositParams v1.DepositParams
	keeper.paramSpace.Get(ctx, v1.ParamStoreKeyDepositParams, &depositParams)
	return depositParams
}

// GetVotingParams returns the current VotingParams from the global param store
func (keeper Keeper) GetVotingParams(ctx sdk.Context) v1.VotingParams {
	var votingParams v1.VotingParams
	keeper.paramSpace.Get(ctx, v1.ParamStoreKeyVotingParams, &votingParams)
	return votingParams
}

// GetTallyParams returns the current TallyParam from the global param store
func (keeper Keeper) GetTallyParams(ctx sdk.Context) v1.TallyParams {
	var tallyParams v1.TallyParams
	keeper.paramSpace.Get(ctx, v1.ParamStoreKeyTallyParams, &tallyParams)
	return tallyParams
}

// SetDepositParams sets DepositParams to the global param store
func (keeper Keeper) SetDepositParams(ctx sdk.Context, depositParams v1.DepositParams) {
	keeper.paramSpace.Set(ctx, v1.ParamStoreKeyDepositParams, &depositParams)
}

// SetVotingParams sets VotingParams to the global param store
func (keeper Keeper) SetVotingParams(ctx sdk.Context, votingParams v1.VotingParams) {
	keeper.paramSpace.Set(ctx, v1.ParamStoreKeyVotingParams, &votingParams)
}

// SetTallyParams sets TallyParams to the global param store
func (keeper Keeper) SetTallyParams(ctx sdk.Context, tallyParams v1.TallyParams) {
	keeper.paramSpace.Set(ctx, v1.ParamStoreKeyTallyParams, &tallyParams)
}
