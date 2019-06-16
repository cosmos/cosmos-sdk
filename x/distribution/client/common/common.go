package common

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// QueryParams actually queries distribution params.
func QueryParams(cliCtx context.CLIContext, queryRoute string) (PrettyParams, error) {
	route := fmt.Sprintf("custom/%s/params/%s", queryRoute, types.ParamCommunityTax)

	retCommunityTax, _, err := cliCtx.QueryWithData(route, []byte{})
	if err != nil {
		return PrettyParams{}, err
	}

	route = fmt.Sprintf("custom/%s/params/%s", queryRoute, types.ParamBaseProposerReward)
	retBaseProposerReward, _, err := cliCtx.QueryWithData(route, []byte{})
	if err != nil {
		return PrettyParams{}, err
	}

	route = fmt.Sprintf("custom/%s/params/%s", queryRoute, types.ParamBonusProposerReward)
	retBonusProposerReward, _, err := cliCtx.QueryWithData(route, []byte{})
	if err != nil {
		return PrettyParams{}, err
	}

	route = fmt.Sprintf("custom/%s/params/%s", queryRoute, types.ParamWithdrawAddrEnabled)
	retWithdrawAddrEnabled, _, err := cliCtx.QueryWithData(route, []byte{})
	if err != nil {
		return PrettyParams{}, err
	}

	return NewPrettyParams(
		retCommunityTax, retBaseProposerReward, retBonusProposerReward, retWithdrawAddrEnabled,
	), nil
}

// QueryDelegatorTotalRewards queries delegator total rewards.
func QueryDelegatorTotalRewards(cliCtx context.CLIContext, queryRoute, delAddr string) ([]byte, error) {
	delegatorAddr, err := sdk.AccAddressFromBech32(delAddr)
	if err != nil {
		return nil, err
	}

	res, _, err := cliCtx.QueryWithData(
		fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegatorTotalRewards),
		cliCtx.Codec.MustMarshalJSON(types.NewQueryDelegatorParams(delegatorAddr)),
	)
	return res, err
}

// QueryDelegationRewards queries a delegation rewards.
func QueryDelegationRewards(cliCtx context.CLIContext, queryRoute, delAddr, valAddr string) ([]byte, error) {
	delegatorAddr, err := sdk.AccAddressFromBech32(delAddr)
	if err != nil {
		return nil, err
	}

	validatorAddr, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return nil, err
	}

	res, _, err := cliCtx.QueryWithData(
		fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegationRewards),
		cliCtx.Codec.MustMarshalJSON(types.NewQueryDelegationRewardsParams(delegatorAddr, validatorAddr)),
	)
	return res, err
}

// QueryDelegatorValidators returns delegator's list of validators
// it submitted delegations to.
func QueryDelegatorValidators(cliCtx context.CLIContext, queryRoute string, delegatorAddr sdk.AccAddress) ([]byte, error) {
	res, _, err := cliCtx.QueryWithData(
		fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegatorValidators),
		cliCtx.Codec.MustMarshalJSON(types.NewQueryDelegatorParams(delegatorAddr)),
	)
	return res, err
}

// QueryValidatorCommission returns a validator's commission.
func QueryValidatorCommission(cliCtx context.CLIContext, queryRoute string, validatorAddr sdk.ValAddress) ([]byte, error) {
	res, _, err := cliCtx.QueryWithData(
		fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidatorCommission),
		cliCtx.Codec.MustMarshalJSON(types.NewQueryValidatorCommissionParams(validatorAddr)),
	)
	return res, err
}

// WithdrawAllDelegatorRewards builds a multi-message slice to be used
// to withdraw all delegations rewards for the given delegator.
func WithdrawAllDelegatorRewards(cliCtx context.CLIContext, queryRoute string, delegatorAddr sdk.AccAddress) ([]sdk.Msg, error) {
	// retrieve the comprehensive list of all validators which the
	// delegator had submitted delegations to
	bz, err := QueryDelegatorValidators(cliCtx, queryRoute, delegatorAddr)
	if err != nil {
		return nil, err
	}

	var validators []sdk.ValAddress
	if err := cliCtx.Codec.UnmarshalJSON(bz, &validators); err != nil {
		return nil, err
	}

	// build multi-message transaction
	var msgs []sdk.Msg
	for _, valAddr := range validators {
		msg := types.NewMsgWithdrawDelegatorReward(delegatorAddr, valAddr)
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
	commissionMsg := types.NewMsgWithdrawValidatorCommission(validatorAddr)
	if err := commissionMsg.ValidateBasic(); err != nil {
		return nil, err
	}

	// build and validate MsgWithdrawDelegatorReward
	rewardMsg := types.NewMsgWithdrawDelegatorReward(sdk.AccAddress(validatorAddr.Bytes()), validatorAddr)
	if err := rewardMsg.ValidateBasic(); err != nil {
		return nil, err
	}

	return []sdk.Msg{commissionMsg, rewardMsg}, nil
}
