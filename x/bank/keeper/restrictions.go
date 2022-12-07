package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// A SendRestrictionFn is a function that can restrict sends and/or provide a new receiver address.
type SendRestrictionFn func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (newToAddr sdk.AccAddress, err error)

// Then creates a composite restriction that runs this one then the provided second one.
func (r SendRestrictionFn) Then(second SendRestrictionFn) SendRestrictionFn {
	return ComposeSendRestrictions(r, second)
}

// ComposeSendRestrictions combines multiple SendRestrictions into one.
// nil entries are ignored.
// If all entries are nil, nil is returned.
// If exactly one entry is not nil, it is returned.
// Otherwise, a new SendRestrictionFn is returned that runs the non-nil restrictions in the order they are given.
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
	return func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
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
