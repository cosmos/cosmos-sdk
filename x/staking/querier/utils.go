package querier

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func delegationToDelegationResponse(ctx sdk.Context, k keeper.Keeper, del types.Delegation) (types.DelegationResponse, sdk.Error) {
	val, found := k.GetValidator(ctx, del.ValidatorAddress)
	if !found {
		return types.DelegationResponse{}, types.ErrNoValidatorFound(types.DefaultCodespace)
	}

	return types.NewDelegationResp(
		del.DelegatorAddress,
		del.ValidatorAddress,
		del.Shares,
		val.TokensFromShares(del.Shares).TruncateInt(),
	), nil
}

func delegationsToDelegationResponses(
	ctx sdk.Context, k keeper.Keeper, delegations types.Delegations,
) (types.DelegationResponses, sdk.Error) {

	resp := make(types.DelegationResponses, len(delegations), len(delegations))
	for i, del := range delegations {
		delResp, err := delegationToDelegationResponse(ctx, k, del)
		if err != nil {
			return nil, err
		}

		resp[i] = delResp
	}

	return resp, nil
}

func redelegationsToRedelegationResponses(
	ctx sdk.Context, k keeper.Keeper, redels types.Redelegations,
) (types.RedelegationResponses, sdk.Error) {

	resp := make(types.RedelegationResponses, len(redels), len(redels))
	for i, redel := range redels {
		val, found := k.GetValidator(ctx, redel.ValidatorDstAddress)
		if !found {
			return nil, types.ErrNoValidatorFound(types.DefaultCodespace)
		}

		entryResponses := make([]types.RedelegationEntryResponse, len(redel.Entries), len(redel.Entries))
		for j, entry := range redel.Entries {
			entryResponses[j] = types.NewRedelegationEntryResponse(
				entry.CreationHeight,
				entry.CompletionTime,
				entry.SharesDst,
				entry.InitialBalance,
				val.TokensFromShares(entry.SharesDst).TruncateInt(),
			)
		}

		resp[i] = types.NewRedelegationResponse(
			redel.DelegatorAddress,
			redel.ValidatorSrcAddress,
			redel.ValidatorDstAddress,
			entryResponses,
		)
	}

	return resp, nil
}
