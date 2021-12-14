package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// GetDepositParams returns the current DepositParams from the global param store
func (keeper Keeper) GetDepositParams(ctx sdk.Context) v1beta1.DepositParams {
	var depositParams v1beta1.DepositParams
	keeper.paramSpace.Get(ctx, types.ParamStoreKeyDepositParams, &depositParams)
	return depositParams
}

// GetVotingParams returns the current VotingParams from the global param store
func (keeper Keeper) GetVotingParams(ctx sdk.Context) v1beta1.VotingParams {
	var votingParams v1beta1.VotingParams
	keeper.paramSpace.Get(ctx, types.ParamStoreKeyVotingParams, &votingParams)
	return votingParams
}

// GetTallyParams returns the current TallyParam from the global param store
func (keeper Keeper) GetTallyParams(ctx sdk.Context) v1beta1.TallyParams {
	var tallyParams v1beta1.TallyParams
	keeper.paramSpace.Get(ctx, types.ParamStoreKeyTallyParams, &tallyParams)
	return tallyParams
}

// SetDepositParams sets DepositParams to the global param store
func (keeper Keeper) SetDepositParams(ctx sdk.Context, depositParams v1beta1.DepositParams) {
	keeper.paramSpace.Set(ctx, types.ParamStoreKeyDepositParams, &depositParams)
}

// SetVotingParams sets VotingParams to the global param store
func (keeper Keeper) SetVotingParams(ctx sdk.Context, votingParams v1beta1.VotingParams) {
	keeper.paramSpace.Set(ctx, types.ParamStoreKeyVotingParams, &votingParams)
}

// SetTallyParams sets TallyParams to the global param store
func (keeper Keeper) SetTallyParams(ctx sdk.Context, tallyParams v1beta1.TallyParams) {
	keeper.paramSpace.Set(ctx, types.ParamStoreKeyTallyParams, &tallyParams)
}
