package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	yaml "gopkg.in/yaml.v2"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName
	// DefaultSendEnabled enabled
	DefaultSendEnabled = true
)

// ParamStoreKeySendEnabledParams is store's key for SendEnabledParams
var ParamStoreKeySendEnabledParams = []byte("sendenabledparams")

// Params is the parameter object for the bank module
type Params struct {
	SendEnabledParams SendEnabledParams `json:"send_enabled_params" yaml:"send_enabled_params"`
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return Params{
		SendEnabledParams: DefaultSendEnabledParams(),
	}
}

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable(
		paramtypes.NewParamSetPair(ParamStoreKeySendEnabledParams, SendEnabledParams{}, validateSendEnabledParams),
	)
}

// SendEnabledParams is a collection of parameters indicating if a coin denom is enabled for sending
type SendEnabledParams []SendEnabledParam

// NewSendEnabledParams creates a new SendEnabledParams object
func NewSendEnabledParams(params ...SendEnabledParam) SendEnabledParams {
	return params
}

// DefaultSendEnabledParams returns the default send enabled parameters
func DefaultSendEnabledParams() SendEnabledParams {
	return SendEnabledParams{
		DefaultSendEnabledParam(),
	}
}

func validateSendEnabledParams(i interface{}) error {
	params, ok := i.(SendEnabledParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	// ensure each denom is only registered one time.
	registered := make(map[string]bool)
	for _, p := range params {
		if err := validateSendEnabledParam(p); err != nil {
			return err
		}
		if _, exists := registered[p.Denom]; exists {
			return fmt.Errorf("duplicate send enabled parameter found: '%s'", p.Denom)
		} else {
			registered[p.Denom] = true
		}
	}
	return nil
}

// Enabled returns true if the given denom is enabled for sending
func (sep SendEnabledParams) Enabled(denom string) bool {
	var defaultSendEnabled = DefaultSendEnabled

	for _, p := range sep {
		if p.Denom == denom {
			return p.SendEnabled
		}
		// capture default case
		if len(p.Denom) == 0 {
			defaultSendEnabled = p.SendEnabled
		}
	}

	return defaultSendEnabled
}

// SetSendEnabledParam returns a SendEnabledParams with the supplied denom/sendEnabled
func (sep SendEnabledParams) SetSendEnabledParam(denom string, sendEnabled bool) SendEnabledParams {
	var results SendEnabledParams
	for _, p := range sep {
		if p.Denom != denom {
			results = append(results, NewSendEnabledParam(p.Denom, p.SendEnabled))
		}

	}
	return append(results, NewSendEnabledParam(denom, sendEnabled))
}

// SendEnabledParams defines if a given denomination (empty for default) is allowed to be sent
type SendEnabledParam struct {
	Denom       string `json:"denom,omitempty" yaml:"denom,omitempty"`
	SendEnabled bool   `json:"send_enabled" yaml:"send_enabled"`
}

// DefaultSendEnabledParams returns the default send enabled parameter
func DefaultSendEnabledParam() SendEnabledParam {
	return SendEnabledParam{
		Denom:       "",
		SendEnabled: true,
	}
}

// NewSendEnabledParam creates a new SendEnabledParams object
func NewSendEnabledParam(denom string, sendEnabled bool) SendEnabledParam {
	return SendEnabledParam{
		Denom:       denom,
		SendEnabled: sendEnabled,
	}
}

// Equal checks equality of SendEnabledParam
func (sep SendEnabledParam) Equal(other SendEnabledParam) bool {
	return strings.EqualFold(sep.Denom, other.Denom) && sep.SendEnabled == other.SendEnabled
}

// String implements stringer insterface
func (sep SendEnabledParam) String() string {
	out, _ := yaml.Marshal(sep)
	return string(out)
}

func validateSendEnabledParam(i interface{}) error {
	param, ok := i.(SendEnabledParam)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	// if the denomination is specified it must be valid
	if len(param.Denom) > 0 {
		return sdk.ValidateDenom(param.Denom)
	}
	return nil
}
