package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"

	abci "github.com/tendermint/tendermint/abci/types"
)

type SendAuthorization struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this authorization and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
}

func (authorization SendAuthorization) MsgType() sdk.Msg {
	return bank.MsgSend{}
}

func (authorization SendAuthorization) Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Authorization, delete bool) {
	switch msg := msg.(type) {
	case bank.MsgSend:
		limitLeft, isNegative := authorization.SpendLimit.SafeSub(msg.Amount)
		if isNegative {
			return false, nil, false
		}
		if limitLeft.IsZero() {
			return true, nil, true
		}

		fmt.Println(limitLeft, "limitLeft")

		return true, SendAuthorization{SpendLimit: limitLeft}, false
	}
	return false, nil, false
}
