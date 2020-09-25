package types

import (
	"fmt"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type Authorization interface {
	proto.Message

	// MsgType returns the type of Msg's that this authorization can accept
	MsgType() string

	// Accept determines whether this grant allows the provided action, and if
	// so provides an upgraded authorization grant
	Accept(msg sdk.Msg, block tmproto.Header) (allow bool, updated Authorization, delete bool)
}

// NewAuthorizationGrant returns new AuthrizationGrant
func NewAuthorizationGrant(authorization Authorization, expiration int64) (*AuthorizationGrant, error) {
	auth := AuthorizationGrant{
		Expiration: expiration,
	}
	msg, ok := authorization.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", authorization)
	}

	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	auth.Authorization = any

	return &auth, nil
}
