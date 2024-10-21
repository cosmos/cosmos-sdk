package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetGrantAuthorization returns the cached value from the Grant.Authorization if present.
func (g Grant) GetGrantAuthorization() (Authorization, error) {
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
