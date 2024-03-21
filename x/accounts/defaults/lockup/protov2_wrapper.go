package lockup

import (
	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	stakingv1beta1 "cosmossdk.io/api/cosmos/staking/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

type ProtoMsg = proto.Message

func makeMsgSend(fromAddr, toAddr string, coins sdk.Coins) ProtoMsg {
	v2Coins := make([]*v1beta1.Coin, len(coins))
	for i, coin := range coins {
		v2Coins[i] = &v1beta1.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.String(),
		}
	}
	return &bankv1beta1.MsgSend{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      v2Coins,
	}
}

func makeMsgDelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) ProtoMsg {
	v2Coin := &v1beta1.Coin{
		Denom:  amount.Denom,
		Amount: amount.Amount.String(),
	}
	return &stakingv1beta1.MsgDelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           v2Coin,
	}
}

func makeMsgUndelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) ProtoMsg {
	v2Coin := &v1beta1.Coin{
		Denom:  amount.Denom,
		Amount: amount.Amount.String(),
	}
	return &stakingv1beta1.MsgUndelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           v2Coin,
	}
}
