package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	yaml "gopkg.in/yaml.v2"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName
)

// KeySendEnabled is store's key for SendEnabled Params
var KeySendEnabled = []byte("SendEnabled")

// ParamKeyTable for bank module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for hte bank module
func NewParams(sendEnabledParams SendEnabledParams) Params {
	return Params{
		SendEnabled: sendEnabledParams,
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return Params{
		SendEnabled: SendEnabledParams{DefaultSendEnabled()},
	}
}

// Validate all bank module parameters
func (p Params) Validate() error {
	if err := validateSendEnabledParams(p.SendEnabled); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// IsSendEnabled returns true if the given denom is enabled for sending
func (p Params) IsSendEnabled(denom string) bool {
	var defaultSendEnabled = true

	for _, p := range p.SendEnabled {
		if p.Denom == denom {
			return p.Enabled
		}
		// capture default case
		if len(p.Denom) == 0 {
			defaultSendEnabled = p.Enabled
		}
	}

	return defaultSendEnabled
}

// SetSendEnabledParam returns an updated set of Parameters with the given denom
// send enabled flag set.
func (p Params) SetSendEnabledParam(denom string, sendEnabled bool) Params {
	var sendParams SendEnabledParams
	for _, p := range p.SendEnabled {
		if p.Denom != denom {
			sendParams = append(sendParams, NewSendEnabled(p.Denom, p.Enabled))
		}
	}
	sendParams = append(sendParams, NewSendEnabled(denom, sendEnabled))
	return NewParams(sendParams)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeySendEnabled, &p.SendEnabled, validateSendEnabledParams),
	}
}

// SendEnabledParams is a collection of parameters indicating if a coin denom is enabled for sending
type SendEnabledParams []*SendEnabled

func validateSendEnabledParams(i interface{}) error {
	params, ok := i.([]*SendEnabled)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	// ensure each denom is only registered one time.
	registered := make(map[string]bool)
	for _, p := range params {
		if err := validateSendEnabled(*p); err != nil {
			return err
		}
		if _, exists := registered[p.Denom]; exists {
			return fmt.Errorf("duplicate send enabled parameter found: '%s'", p.Denom)
		}
		registered[p.Denom] = true
	}
	return nil
}

// DefaultSendEnabled returns the default send enabled parameter
func DefaultSendEnabled() *SendEnabled {
	return &SendEnabled{
		Denom:   "",
		Enabled: true,
	}
}

// NewSendEnabled creates a new SendEnabled object
func NewSendEnabled(denom string, sendEnabled bool) *SendEnabled {
	return &SendEnabled{
		Denom:   denom,
		Enabled: sendEnabled,
	}
}

// String implements stringer insterface
func (se SendEnabled) String() string {
	out, _ := yaml.Marshal(se)
	return string(out)
}

func validateSendEnabled(i interface{}) error {
	param, ok := i.(SendEnabled)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	// if the denomination is specified it must be valid
	if len(param.Denom) > 0 {
		return sdk.ValidateDenom(param.Denom)
	}
	return nil
}
