package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A SendRestrictionFn can restrict sends and/or provide a new receiver address.
type SendRestrictionFn func(ctx context.Context, fromAddr, toAddr []byte, amt sdk.Coins) (newToAddr []byte, err error)

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (SendRestrictionFn) IsOnePerModuleType() {}

var _ SendRestrictionFn = NoOpSendRestrictionFn

// NoOpSendRestrictionFn is a no-op SendRestrictionFn.
func NoOpSendRestrictionFn(_ context.Context, _, toAddr []byte, _ sdk.Coins) ([]byte, error) {
	return toAddr, nil
}

// Then creates a composite restriction that runs this one then the provided second one.
func (r SendRestrictionFn) Then(second SendRestrictionFn) SendRestrictionFn {
	return ComposeSendRestrictions(r, second)
}

// ComposeSendRestrictions combines multiple SendRestrictionFn into one.
// nil entries are ignored.
// If all entries are nil, nil is returned.
// If exactly one entry is not nil, it is returned.
// Otherwise, a new SendRestrictionFn is returned that runs the non-nil restrictions in the order they are given.
// The composition runs each send restriction until an error is encountered and returns that error,
// otherwise it returns the toAddr of the last send restriction.
func ComposeSendRestrictions(restrictions ...SendRestrictionFn) SendRestrictionFn {
	toRun := make([]SendRestrictionFn, 0, len(restrictions))
	for _, r := range restrictions {
		if r != nil {
			toRun = append(toRun, r)
		}
	}
	switch len(toRun) {
	case 0:
		return nil
	case 1:
		return toRun[0]
	}
	return func(ctx context.Context, fromAddr, toAddr []byte, amt sdk.Coins) ([]byte, error) {
		var err error
		for _, r := range toRun {
			toAddr, err = r(ctx, fromAddr, toAddr, amt)
			if err != nil {
				return toAddr, err
			}
		}
		return toAddr, err
	}
}
