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

// QueryDelegatorTotalRewards queries delegator total rewards.
func QueryDelegatorTotalRewards(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute, delAddr string) ([]byte, error) {

	delegatorAddr, err := sdk.AccAddressFromBech32(delAddr)
	if err != nil {
		return nil, err
	}

	return cliCtx.QueryWithData(
		fmt.Sprintf("custom/%s/delegator_total_rewards", queryRoute),
		cdc.MustMarshalJSON(distr.NewQueryDelegatorParams(delegatorAddr)),
	)
}

// QueryDelegationRewards queries a delegation rewards.
func QueryDelegationRewards(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute, delAddr, valAddr string) ([]byte, error) {

	delegatorAddr, err := sdk.AccAddressFromBech32(delAddr)
	if err != nil {
		return nil, err
	}
	validatorAddr, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return nil, err
	}

	return cliCtx.QueryWithData(
		fmt.Sprintf("custom/%s/delegation_rewards", queryRoute),
		cdc.MustMarshalJSON(distr.NewQueryDelegationRewardsParams(delegatorAddr, validatorAddr)),
	)
}

// QueryDelegatorValidators returns delegator's list of validators
// it submitted delegations to.
func QueryDelegatorValidators(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string, delegatorAddr sdk.AccAddress) ([]byte, error) {

	return cliCtx.QueryWithData(
		fmt.Sprintf("custom/%s/delegator_validators", queryRoute),
		cdc.MustMarshalJSON(distr.NewQueryDelegatorParams(delegatorAddr)),
	)
}

// QueryValidatorCommission returns a validator's commission.
func QueryValidatorCommission(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string, validatorAddr sdk.ValAddress) ([]byte, error) {

	return cliCtx.QueryWithData(
		fmt.Sprintf("custom/%s/validator_commission", queryRoute),
		cdc.MustMarshalJSON(distr.NewQueryValidatorCommissionParams(validatorAddr)),
	)
}

// WithdrawAllDelegatorRewards builds a multi-message slice to be used
// to withdraw all delegations rewards for the given delegator.
func WithdrawAllDelegatorRewards(cliCtx context.CLIContext, cdc *codec.Codec,
	queryRoute string, delegatorAddr sdk.AccAddress) ([]sdk.Msg, error) {

	// retrieve the comprehensive list of all validators which the
	// delegator had submitted delegations to
	bz, err := QueryDelegatorValidators(cliCtx, cdc, queryRoute, delegatorAddr)
	if err != nil {
		return nil, err
	}

	var validators []sdk.ValAddress
	if err := cdc.UnmarshalJSON(bz, &validators); err != nil {
		return nil, err
	}

	// build multi-message transaction
	var msgs []sdk.Msg
	for _, valAddr := range validators {
		msg := distr.NewMsgWithdrawDelegatorReward(delegatorAddr, valAddr)
		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
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
