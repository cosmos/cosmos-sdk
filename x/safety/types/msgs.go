package types

import (
	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgTrip{}
	_ sdk.Msg = &MsgReset{}
)

// NewMsgTrip creates a new MsgTrip.
func NewMsgTrip(operator string, patterns ...string) *MsgTrip {
	return &MsgTrip{Operator: operator, Patterns: patterns}
}

// NewMsgReset creates a new MsgReset.
func NewMsgReset(sender string, patterns ...string) *MsgReset {
	return &MsgReset{Sender: sender, Patterns: patterns}
}

// Validate performs stateless checks on MsgUpdateParams.
func (m *MsgUpdateParams) Validate(ac address.Codec) error {
	if err := validateAddress(ac, "authority", m.Authority); err != nil {
		return err
	}
	return m.Params.Validate(ac)
}

// Validate performs stateless checks on MsgTrip. The operator must be in
// canonical form, as params.operator is, so EventTrip names the operator
// exactly as params does.
func (m *MsgTrip) Validate(ac address.Codec) error {
	if m.Operator == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("operator must be set")
	}
	if err := ValidateCanonicalAddress(ac, m.Operator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("operator %q: %s", m.Operator, err)
	}
	return ValidatePatterns(m.Patterns)
}

// Validate performs stateless checks on MsgReset.
func (m *MsgReset) Validate(ac address.Codec) error {
	if err := validateAddress(ac, "sender", m.Sender); err != nil {
		return err
	}
	return ValidatePatterns(m.Patterns)
}

func validateAddress(ac address.Codec, field, addr string) error {
	if addr == "" {
		return sdkerrors.ErrInvalidAddress.Wrapf("%s must be set", field)
	}
	if _, err := ac.StringToBytes(addr); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("%s %q: %s", field, addr, err)
	}
	return nil
}
