package lockup

import (
	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	stakingv1beta1 "cosmossdk.io/api/cosmos/staking/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
)

type ProtoMsg = protoiface.MessageV1

type gogoProtoPlusV2 interface {
	proto.Message
	ProtoMsg
}

// protoV2GogoWrapper is a wrapper of a protov2 message into a gogo message.
// this is exceptionally allowed to enable accounts to be decoupled from
// the SDK, since x/accounts can support only protov1 in its APIs.
// But in order to keep it decoupled from the SDK we need to use the API module.
// This is a temporary solution that is being used here:
// https://github.com/cosmos/cosmos-sdk/blob/main/x/accounts/coin_transfer.go
type protoV2GogoWrapper struct {
	gogoProtoPlusV2
}

func (h protoV2GogoWrapper) XXX_MessageName() string {
	return string(proto.MessageName(h.gogoProtoPlusV2))
}

func makeMsgSend(fromAddr, toAddr string, coins sdk.Coins) ProtoMsg {
	v2Coins := make([]*v1beta1.Coin, len(coins))
	for i, coin := range coins {
		v2Coins[i] = &v1beta1.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.String(),
		}
	}
	return protoV2GogoWrapper{&bankv1beta1.MsgSend{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      v2Coins,
	}}
}

func makeMsgDelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) ProtoMsg {
	v2Coin := &v1beta1.Coin{
		Denom:  amount.Denom,
		Amount: amount.Amount.String(),
	}
	return protoV2GogoWrapper{&stakingv1beta1.MsgDelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           v2Coin,
	}}
}

func makeMsgUndelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) ProtoMsg {
	v2Coin := &v1beta1.Coin{
		Denom:  amount.Denom,
		Amount: amount.Amount.String(),
	}
	return protoV2GogoWrapper{&stakingv1beta1.MsgUndelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           v2Coin,
	}}
}
