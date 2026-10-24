package types

import (
	"fmt"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultParams returns params with the given authority, no operator, and an
// empty scope, so an operator can do nothing until the authority grants scope.
func DefaultParams(authority string) Params {
	return Params{
		Authority: authority,
		Operator:  "",
	}
}

// Validate checks that the authority is a valid address, the operator is empty
// or a valid address, and the scope patterns are well-formed and unique.
// Addresses must be in canonical form, because sdk.ValidateAuthority compares
// strings: an authority stored as an uppercase bech32 string would reject a
// msg signed by the same account in its usual lowercase form.
func (p Params) Validate(ac address.Codec) error {
	if p.Authority == "" {
		return ErrInvalidParams.Wrap("authority must be set")
	}
	if err := ValidateCanonicalAddress(ac, p.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("authority %q: %s", p.Authority, err)
	}
	if p.Operator != "" {
		if err := ValidateCanonicalAddress(ac, p.Operator); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("operator %q: %s", p.Operator, err)
		}
	}
	if err := validateUniquePatterns(p.Scope); err != nil {
		return errorsmod.Wrap(err, "scope")
	}
	return nil
}

// ValidateCanonicalAddress checks that addr decodes and re-encodes to itself.
func ValidateCanonicalAddress(ac address.Codec, addr string) error {
	bz, err := ac.StringToBytes(addr)
	if err != nil {
		return err
	}
	canonical, err := ac.BytesToString(bz)
	if err != nil {
		return err
	}
	if canonical != addr {
		return fmt.Errorf("not in canonical form, expected %q", canonical)
	}
	return nil
}
