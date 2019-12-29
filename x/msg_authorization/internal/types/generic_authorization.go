package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GenericAuthorization grants the permission to execute any transaction of the provided
// sdk.Msg type without restrictions
type GenericAuthorization struct {
	// MsgType is the type of Msg this capability grant allows
	Msg sdk.Msg
}

func (cap GenericAuthorization) MsgType() sdk.Msg {
	return cap.MsgType()
}

func (cap GenericAuthorization) Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Authorization, delete bool) {
	return true, cap, false
}
