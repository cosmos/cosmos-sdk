package lockup

import (
	"github.com/cosmos/gogoproto/proto"

	banktypes "cosmossdk.io/x/bank/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: this file will become into "utils" or something like that, or maybe we remove it

type ProtoMsg = proto.Message

func makeMsgSend(fromAddr, toAddr string, coins sdk.Coins) ProtoMsg {
	return &banktypes.MsgSend{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      coins,
	}
}

func makeMsgDelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) ProtoMsg {
	return &stakingtypes.MsgDelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           amount,
	}
}

func makeMsgUndelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) ProtoMsg {
	return &stakingtypes.MsgUndelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           amount,
	}
}
