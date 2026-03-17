package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ValidateAuthority checks that msgAuthority matches the effective authority.
// It first checks consensus params; if no authority is set there, it falls back
// to keeperAuthority. Returns nil on success, or an ErrUnauthorized error if mismatched.
// If the keeper has no authority, use an empty string ("") as input.
func ValidateAuthority(ctx Context, keeperAuthority, msgAuthority string) error {
	expected := keeperAuthority
	if cp := ctx.ConsensusParams(); cp.Authority != nil && cp.Authority.Authority != "" {
		expected = cp.Authority.Authority
	}
	if expected != msgAuthority {
		return sdkerrors.ErrUnauthorized.Wrapf("invalid authority: expected %s, got %s", expected, msgAuthority)
	}
	return nil
}
