package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GenericCapability grants the permission to execute any transaction of the provided
// sdk.Msg type without restrictions
type GenericCapability struct {
	// MsgType is the type of Msg this capability grant allows
	Msg sdk.Msg
}

func (cap GenericCapability) MsgType() sdk.Msg {
	return cap.MsgType()
}

func (cap GenericCapability) Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Capability, delete bool) {
	return true, cap, false
}
