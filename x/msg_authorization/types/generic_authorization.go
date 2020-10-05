package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	_ Authorization = &GenericAuthorization{}
)

func NewGenericAuthorization (msgType string) *GenericAuthorization {
	// var msg sdk.Msg
	// TODO handle generic msg type
	return nil
}

func (cap GenericAuthorization) MsgType() string {
	var msg sdk.Msg
	ModuleCdc.UnpackAny(cap.Message, &msg)
	return msg.Type()
}

func (cap GenericAuthorization) Accept(msg sdk.Msg, block tmproto.Header) (allow bool, updated Authorization, delete bool) {
	return true, &cap, false
}
