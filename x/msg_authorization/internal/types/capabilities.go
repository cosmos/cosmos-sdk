package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	abci "github.com/tendermint/tendermint/abci/types"
)

type Capability interface {
	MsgType() sdk.Msg
	Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Capability, delete bool)
}

type SendCapability struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this capability and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
}

func (capability SendCapability) MsgType() sdk.Msg {
	return bank.MsgSend{}
}

func (capability SendCapability) Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Capability, delete bool) {
	switch msg := msg.(type) {
	case bank.MsgSend:
		limitLeft, isNegative := capability.SpendLimit.SafeSub(msg.Amount)
		if isNegative {
			return false, nil, false
		}
		if limitLeft.IsZero() {
			return true, nil, true
		}
		return true, SendCapability{SpendLimit: limitLeft}, false
	}
	return false, nil, false
}
