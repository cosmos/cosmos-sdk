package common

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
)

// QueryParams actually queries distribution params.
func QueryParams(cliCtx context.CLIContext, queryRoute string) (PrettyParams, error) {
	route := fmt.Sprintf("custom/%s/params/community_tax", queryRoute)

	retCommunityTax, err := cliCtx.QueryWithData(route, []byte{})
	if err != nil {
		return PrettyParams{}, err
	}

	route = fmt.Sprintf("custom/%s/params/base_proposer_reward", queryRoute)
	retBaseProposerReward, err := cliCtx.QueryWithData(route, []byte{})
	if err != nil {
		return PrettyParams{}, err
	}

	route = fmt.Sprintf("custom/%s/params/bonus_proposer_reward", queryRoute)
	retBonusProposerReward, err := cliCtx.QueryWithData(route, []byte{})
	if err != nil {
		return PrettyParams{}, err
	}

	route = fmt.Sprintf("custom/%s/params/withdraw_addr_enabled", queryRoute)
	retWithdrawAddrEnabled, err := cliCtx.QueryWithData(route, []byte{})
	if err != nil {
		return PrettyParams{}, err
	}

	return NewPrettyParams(retCommunityTax, retBaseProposerReward,
		retBonusProposerReward, retWithdrawAddrEnabled), nil
}

// QueryParams queries delegator rewards. If valAddr is empty string,
// it returns all delegations rewards for the given delegator; else
// it returns the rewards for the specific delegation.
func QueryRewards(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute, delAddr, valAddr string) ([]byte, error) {

	delegatorAddr, err := sdk.AccAddressFromBech32(delAddr)
	if err != nil {
		return nil, err
	}

	var params distr.QueryDelegationRewardsParams
	var route string

	if valAddr == "" {
		params = distr.NewQueryDelegationRewardsParams(delegatorAddr, nil)
		route = fmt.Sprintf("custom/%s/all_delegation_rewards", queryRoute)
	} else {
		validatorAddr, err := sdk.ValAddressFromBech32(valAddr)
		if err != nil {
			return nil, err
		}

		params = distr.NewQueryDelegationRewardsParams(delegatorAddr, validatorAddr)
		route = fmt.Sprintf("custom/%s/delegation_rewards", queryRoute)
	}

	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return nil, err
	}

	return cliCtx.QueryWithData(route, bz)
}

// QueryValidatorCommission returns a validator's commission.
func QueryValidatorCommission(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string, validatorAddr sdk.ValAddress) ([]byte, error) {

	return cliCtx.QueryWithData(
		fmt.Sprintf("custom/%s/validator_commission", queryRoute),
		cdc.MustMarshalJSON(distr.NewQueryValidatorCommissionParams(validatorAddr)),
	)
}

// WithdrawValidatorRewardsAndCommission builds a two-message message slice to be
// used to withdraw both validation's commission and self-delegation reward.
func WithdrawValidatorRewardsAndCommission(validatorAddr sdk.ValAddress) ([]sdk.Msg, error) {

	commissionMsg := distr.NewMsgWithdrawValidatorCommission(validatorAddr)
	if err := commissionMsg.ValidateBasic(); err != nil {
		return nil, err
	}

	// build and validate MsgWithdrawDelegatorReward
	rewardMsg := distr.NewMsgWithdrawDelegatorReward(
		sdk.AccAddress(validatorAddr.Bytes()), validatorAddr)
	if err := rewardMsg.ValidateBasic(); err != nil {
		return nil, err
	}

	return []sdk.Msg{commissionMsg, rewardMsg}, nil
}
