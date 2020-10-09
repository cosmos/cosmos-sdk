package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	_ Authorization = &GenericAuthorization{}
)

func (cap GenericAuthorization) MsgType() string {
	var msg sdk.Msg
	ModuleCdc.UnpackAny(cap.Message, &msg)
	return proto.MessageName(msg)
}

func (cap GenericAuthorization) Accept(msg sdk.Msg, block tmproto.Header) (allow bool, updated Authorization, delete bool) {
	return true, &cap, false
}
