package types

import (
	"time"

	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// NewGrant returns new AuthrizationGrant
func NewGrant(authorization authz.Authorization, expiration time.Time) (Grant, error) {
	auth := Grant{
		Expiration: expiration,
	}
	msg, ok := authorization.(proto.Message)
	if !ok {
		return Grant{}, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", authorization)
	}

	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return Grant{}, err
	}

	auth.Authorization = any

	return auth, nil
}

var (
	_ types.UnpackInterfacesMessage = &Grant{}
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (auth Grant) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var authorization authz.Authorization
	return unpacker.UnpackAny(auth.Authorization, &authorization)
}

// GetAuthorization returns the cached value from the Grant.Authorization if present.
func (auth Grant) GetAuthorization() authz.Authorization {
	authorization, ok := auth.Authorization.GetCachedValue().(authz.Authorization)
	if !ok {
		return nil
	}
	return authorization
}
