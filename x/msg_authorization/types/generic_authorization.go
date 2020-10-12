package types

import (
	"fmt"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	_ Authorization = &GenericAuthorization{}
)

func NewGenericAuthorization(msg sdk.Msg) (GenericAuthorization, error) {
	auth := GenericAuthorization{}

	msg1, ok := msg.(proto.Message)
	if !ok {
		return GenericAuthorization{}, fmt.Errorf("cannot proto marshal %T", msg)
	}

	any, err := types.NewAnyWithValue(msg1)
	if err != nil {
		return GenericAuthorization{}, err
	}

	auth.Message = any
	return auth, nil
}

func (cap GenericAuthorization) MsgType() string {
	var msg sdk.Msg
	ModuleCdc.UnpackAny(cap.Message, &msg)
	return proto.MessageName(msg)
}

func (cap GenericAuthorization) Accept(msg sdk.Msg, block tmproto.Header) (allow bool, updated Authorization, delete bool) {
	return true, &cap, false
}
