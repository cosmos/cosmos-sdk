package authz

import (
	"time"

	proto "github.com/gogo/protobuf/proto"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGrant returns new AuthrizationGrant
func NewGrant(authorization Authorization, expiration time.Time) (Grant, error) {
	auth := Grant{
		Expiration: expiration,
	}
	msg, ok := authorization.(proto.Message)
	if !ok {
		return Grant{}, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", authorization)
	}

	any, err := cdctypes.NewAnyWithValue(msg)
	if err != nil {
		return Grant{}, err
	}

	auth.Authorization = any

	return auth, nil
}

var (
	_ cdctypes.UnpackInterfacesMessage = &Grant{}
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (auth Grant) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var authorization Authorization
	return unpacker.UnpackAny(auth.Authorization, &authorization)
}

// GetAuthorization returns the cached value from the Grant.Authorization if present.
func (auth Grant) GetAuthorization() Authorization {
	authorization, ok := auth.Authorization.GetCachedValue().(Authorization)
	if !ok {
		return nil
	}
	return authorization
}
