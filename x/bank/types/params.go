package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultDefaultSendEnabled is the value that DefaultSendEnabled will have from DefaultParams().
var DefaultDefaultSendEnabled = true

// NewParams creates a new parameter configuration for the bank module
func NewParams(defaultSendEnabled bool) Params {
	return Params{
		SendEnabled:        nil,
		DefaultSendEnabled: defaultSendEnabled,
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return Params{
		SendEnabled:        nil,
		DefaultSendEnabled: DefaultDefaultSendEnabled,
	}
}

// Validate all bank module parameters
func (p Params) Validate() error {
	if len(p.SendEnabled) > 0 {
		return errors.New("use of send_enabled in params is no longer supported")
	}
	return nil
}

// Validate gets any errors with this SendEnabled entry.
func (se SendEnabled) Validate() error {
	return sdk.ValidateDenom(se.Denom)
}

// NewSendEnabled creates a new SendEnabled object
// The denom may be left empty to control the global default setting of send_enabled
func NewSendEnabled(denom string, sendEnabled bool) *SendEnabled {
	return &SendEnabled{
		Denom:   denom,
		Enabled: sendEnabled,
	}
}
