package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CreateCovenantMessage struct {
	Sender    sdk.Address
	Settlers  []sdk.Address
	Receivers []sdk.Address
	Amount    sdk.Coins
}

type SettleCovenantMessage struct {
	CovID    int64
	Settler  sdk.Address
	Receiver sdk.Address
}
