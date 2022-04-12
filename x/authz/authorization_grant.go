package authz

import (
	"time"

	proto "github.com/gogo/protobuf/proto"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGrant returns new Grant. Expiration is optional and noop if null.
// It returns an error if the expiration is before the current block time,
// which is passed into the `blockTime` arg.
func NewGrant(blockTime time.Time, a Authorization, expiration *time.Time) (Grant, error) {
	if expiration != nil && !expiration.After(blockTime) {
		return Grant{}, sdkerrors.ErrInvalidRequest.Wrapf("expiration must be after the current block time (%v), got %v", blockTime.Format(time.RFC3339), expiration.Format(time.RFC3339))
	}
	msg, ok := a.(proto.Message)
	if !ok {
		return Grant{}, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", a)
	}
	any, err := cdctypes.NewAnyWithValue(msg)
	if err != nil {
		return Grant{}, err
	}
	return Grant{
		Expiration:    expiration,
		Authorization: any,
	}, nil
}

var (
	_ cdctypes.UnpackInterfacesMessage = &Grant{}
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (g Grant) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var authorization Authorization
	return unpacker.UnpackAny(g.Authorization, &authorization)
}

// GetAuthorization returns the cached value from the Grant.Authorization if present.
func (g Grant) GetAuthorization() (Authorization, error) {
	if g.Authorization == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "authorization is nil")
	}
	a, ok := g.Authorization.GetCachedValue().(Authorization)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("expected %T, got %T", (Authorization)(nil), g.Authorization.GetCachedValue())
	}
	return a, nil
}

func (g Grant) ValidateBasic() error {
	av := g.Authorization.GetCachedValue()
	a, ok := av.(Authorization)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", (Authorization)(nil), av)
	}
	return a.ValidateBasic()
}
