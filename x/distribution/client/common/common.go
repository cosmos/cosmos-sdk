package common

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// QueryDelegationRewards queries a delegation rewards between a delegator and a
// validator.
func QueryDelegationRewards(clientCtx client.Context, queryRoute, delAddr, valAddr string) ([]byte, int64, error) {
	delegatorAddr, err := sdk.AccAddressFromBech32(delAddr)
	if err != nil {
		return nil, 0, err
	}

	validatorAddr, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return nil, 0, err
	}

	params := types.NewQueryDelegationRewardsParams(delegatorAddr, validatorAddr)
	bz, err := clientCtx.Codec.MarshalJSON(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegationRewards)
	return clientCtx.QueryWithData(route, bz)
}

// QueryDelegatorValidators returns delegator's list of validators
// it submitted delegations to.
func QueryDelegatorValidators(clientCtx client.Context, queryRoute string, delegatorAddr sdk.AccAddress) ([]byte, error) {
	res, _, err := clientCtx.QueryWithData(
		fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegatorValidators),
		clientCtx.Codec.MustMarshalJSON(types.NewQueryDelegatorParams(delegatorAddr)),
	)
	return res, err
}

// QueryValidatorCommission returns a validator's commission.
func QueryValidatorCommission(clientCtx client.Context, queryRoute string, validatorAddr sdk.ValAddress) ([]byte, error) {
	res, _, err := clientCtx.QueryWithData(
		fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidatorCommission),
		clientCtx.Codec.MustMarshalJSON(types.NewQueryValidatorCommissionParams(validatorAddr)),
	)
	return res, err
}

// WithdrawAllDelegatorRewards builds a multi-message slice to be used
// to withdraw all delegations rewards for the given delegator.
func WithdrawAllDelegatorRewards(clientCtx client.Context, queryRoute string, delegatorAddr sdk.AccAddress) ([]sdk.Msg, error) {
	// retrieve the comprehensive list of all validators which the
	// delegator had submitted delegations to
	bz, err := QueryDelegatorValidators(clientCtx, queryRoute, delegatorAddr)
	if err != nil {
		return nil, err
	}

	var validators []sdk.ValAddress
	if err := clientCtx.Codec.UnmarshalJSON(bz, &validators); err != nil {
		return nil, err
	}

	// build multi-message transaction
	msgs := make([]sdk.Msg, 0, len(validators))
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
