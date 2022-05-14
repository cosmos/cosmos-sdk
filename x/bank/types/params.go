package types

import (
	"errors"
	"fmt"

	"sigs.k8s.io/yaml"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// DefaultDefaultSendEnabled is the value that DefaultSendEnabled will have from DefaultParams().
	DefaultDefaultSendEnabled = true
)

var (
	// KeySendEnabled is store's key for SendEnabled Params
	// Deprecated: Use the SendEnabled functionality in the keeper.
	KeySendEnabled = []byte("SendEnabled")
	// KeyDefaultSendEnabled is store's key for the DefaultSendEnabled option
	KeyDefaultSendEnabled = []byte("DefaultSendEnabled")
)

// ParamKeyTable for bank module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for the bank module
func NewParams(defaultSendEnabled bool) Params {
	return Params{
		SendEnabled:        []*SendEnabled{},
		DefaultSendEnabled: defaultSendEnabled,
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return Params{
		SendEnabled:        []*SendEnabled{},
		DefaultSendEnabled: DefaultDefaultSendEnabled,
	}
}

// Validate all bank module parameters
func (p Params) Validate() error {
	if len(p.SendEnabled) > 0 {
		return errors.New("use of send_enabled in params is no longer supported")
	}
	return validateIsBool(p.DefaultSendEnabled)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeySendEnabled, &p.SendEnabled, validateSendEnabledParams),
		paramtypes.NewParamSetPair(KeyDefaultSendEnabled, &p.DefaultSendEnabled, validateIsBool),
	}
}

func (se SendEnabled) Validate() error {
	return sdk.ValidateDenom(se.Denom)
}

func validateSendEnabledParams(i interface{}) error {
	params, ok := i.([]*SendEnabled)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(params) > 0 {
		return errors.New("use of send_enabled in params is no longer supported")
	}
	return nil
}

// NewSendEnabled creates a new SendEnabled object
// The denom may be left empty to control the global default setting of send_enabled
func NewSendEnabled(denom string, sendEnabled bool) *SendEnabled {
	return &SendEnabled{
		Denom:   denom,
		Enabled: sendEnabled,
	}
}

// String implements stringer interface
func (se SendEnabled) String() string {
	return fmt.Sprintf("%q: %t", se.Denom, se.Enabled)
}

func validateSendEnabled(i interface{}) error {
	param, ok := i.(SendEnabled)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return param.Validate()
}

func validateIsBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
