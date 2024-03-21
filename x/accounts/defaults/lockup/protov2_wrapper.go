package lockup

import (
	"github.com/cosmos/gogoproto/proto"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	stakingv1beta1 "cosmossdk.io/api/cosmos/staking/v1beta1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: this file will become into "utils" or something like that, or maybe we remove it

type ProtoMsg = proto.Message

func makeMsgSend(fromAddr, toAddr string, coins sdk.Coins) ProtoMsg {
	return &bankv1beta1.MsgSend{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      coins,
	}
}

func makeMsgDelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) ProtoMsg {
	v2Coin := v1beta1.Coin{
		Denom:  amount.Denom,
		Amount: amount.Amount,
	}
	return &stakingv1beta1.MsgDelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           v2Coin,
	}
}

func makeMsgUndelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) ProtoMsg {
	v2Coin := v1beta1.Coin{
		Denom:  amount.Denom,
		Amount: amount.Amount,
	}
	return &stakingv1beta1.MsgUndelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           v2Coin,
	}
}
