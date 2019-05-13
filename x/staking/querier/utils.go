package querier

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func delegationToDelegationResp(ctx sdk.Context, k keeper.Keeper, del types.Delegation) (types.DelegationResp, sdk.Error) {
	val, found := k.GetValidator(ctx, del.ValidatorAddress)
	if !found {
		return types.DelegationResp{}, types.ErrNoValidatorFound(types.DefaultCodespace)
	}

	return types.NewDelegationResp(
		del.DelegatorAddress,
		del.ValidatorAddress,
		val.TokensFromShares(del.Shares).TruncateInt(),
	), nil
}

func delegationsToDelegationResps(
	ctx sdk.Context, k keeper.Keeper, delegations types.Delegations,
) (types.DelegationResponses, sdk.Error) {

	resp := make(types.DelegationResponses, 0, len(delegations))
	for _, del := range delegations {
		delResp, err := delegationToDelegationResp(ctx, k, del)
		if err != nil {
			return nil, err
		}

		resp = append(resp, delResp)
	}

	return resp, nil
}
