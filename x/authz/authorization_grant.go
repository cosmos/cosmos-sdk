package authz

import (
	"time"

	"github.com/cosmos/gogoproto/proto"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"

	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGrant returns new Grant
func NewGrant( /*blockTime time.Time, */ a Authorization, expiration time.Time) (Grant, error) {
	// TODO: add this for 0.45
	// if !expiration.After(blockTime) {
	// 	return Grant{}, sdkerrors.ErrInvalidRequest.Wrapf("expiration must be after the current block time (%v), got %v", blockTime.Format(time.RFC3339), expiration.Format(time.RFC3339))
	// }
	g := Grant{
		Expiration: expiration,
	}
	msg, ok := a.(proto.Message)
	if !ok {
		return Grant{}, sdkerrors.ErrPackAny.Wrapf("cannot proto marshal %T", a)
	}
	any, err := gogoprotoany.NewAnyWithCacheWithValue(msg)
	if err != nil {
		return Grant{}, err
	}
	return Grant{
		Expiration:    expiration,
		Authorization: any,
	}, nil
}

var _ gogoprotoany.UnpackInterfacesMessage = &Grant{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (g Grant) UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error {
	var authorization Authorization
	return unpacker.UnpackAny(g.Authorization, &authorization)
}

// GetAuthorization returns the cached value from the Grant.Authorization if present.
func (g Grant) GetAuthorization() (Authorization, error) {
	if g.Authorization == nil {
		return nil, sdkerrors.ErrInvalidType.Wrap("authorization is nil")
	}
	av := g.Authorization.GetCachedValue()
	a, ok := av.(Authorization)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("expected %T, got %T", (Authorization)(nil), av)
	}
	return a, nil
}

func (g Grant) ValidateBasic() error {
	av := g.Authorization.GetCachedValue()
	a, ok := av.(Authorization)
	if !ok {
		return sdkerrors.ErrInvalidType.Wrapf("expected %T, got %T", (Authorization)(nil), av)
	}
	return a.ValidateBasic()
}
