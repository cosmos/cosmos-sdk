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

	// MethodName returns the fully-qualified Msg service method name as described in ADR 031.
	MethodName() string

	// Accept determines whether this grant permits the provided sdk.ServiceMsg to be performed, and if
	// so provides an upgraded authorization instance.
	Accept(msg sdk.ServiceMsg, block tmproto.Header) (allow bool, updated Authorization, delete bool)
}

// NewAuthorizationGrant returns new AuthrizationGrant
func NewAuthorizationGrant(authorization Authorization, expiration int64) (AuthorizationGrant, error) {
	auth := AuthorizationGrant{
		Expiration: expiration,
	}
	msg, ok := authorization.(proto.Message)
	if !ok {
		return AuthorizationGrant{}, fmt.Errorf("cannot proto marshal %T", authorization)
	}

	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return AuthorizationGrant{}, err
	}

	auth.Authorization = any

	return auth, nil
}

var (
	_ types.UnpackInterfacesMessage = &AuthorizationGrant{}
)

func (auth AuthorizationGrant) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var authorization Authorization
	return unpacker.UnpackAny(auth.Authorization, &authorization)
}

func (auth AuthorizationGrant) GetAuthorization() Authorization {
	authorization, ok := auth.Authorization.GetCachedValue().(Authorization)
	if !ok {
		return nil
	}
	return authorization
}
