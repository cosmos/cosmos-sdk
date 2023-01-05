package types

import (
	"github.com/cockroachdb/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgAuthorizeCircuitBreaker{}
	_ sdk.Msg = &MsgTripCircuitBreaker{}
	_ sdk.Msg = &MsgResetCircuitBreaker{}
)

// Route Implements Msg.
func (m MsgAuthorizeCircuitBreaker) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgAuthorizeCircuitBreaker) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgAuthorizeCircuitBreaker) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgAuthorizeCircuitBreaker.
func (m MsgAuthorizeCircuitBreaker) GetSigners() []sdk.AccAddress {
	granter := sdk.MustAccAddressFromBech32(m.Granter)

	return []sdk.AccAddress{granter}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgAuthorizeCircuitBreaker) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Granter)
	if err != nil {
		return errors.Wrap(err, "granter")
	}

	_, err = sdk.AccAddressFromBech32(m.Grantee)
	if err != nil {
		return errors.Wrap(err, "granter")
	}

	return nil
}

// Route Implements Msg.
func (m MsgTripCircuitBreaker) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgTripCircuitBreaker) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgTripCircuitBreaker) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgTripCircuitBreaker.
func (m MsgTripCircuitBreaker) GetSigners() []sdk.AccAddress {
	granter := sdk.MustAccAddressFromBech32(m.Authority)

	return []sdk.AccAddress{granter}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgTripCircuitBreaker) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Wrap(err, "granter")
	}

	return nil
}

// Route Implements Msg.
func (m MsgResetCircuitBreaker) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgResetCircuitBreaker) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgResetCircuitBreaker) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgResetCircuitBreaker.
func (m MsgResetCircuitBreaker) GetSigners() []sdk.AccAddress {
	granter := sdk.MustAccAddressFromBech32(m.Authority)

	return []sdk.AccAddress{granter}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgResetCircuitBreaker) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Wrap(err, "granter")
	}

	return nil
}
