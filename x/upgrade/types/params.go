package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyIsMainnet = []byte("IsMainnet")
)

// ParamKeyTable returns a KeyTable for the x/upgrade module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(isMainnet bool) Params {
	return Params{
		IsMainnet: isMainnet,
	}
}

// DefaultParams returns the x/upgrade module's default parameters.
func DefaultParams() Params {
	return NewParams(true)
}

// ParamSetPairs returns the ParamSetPairs for the x/upgrade module.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyIsMainnet, &p.IsMainnet, validateIsMainnet),
	}
}

// Validate performs basic validation on the Params type returning an error upon
// validation failure.
func (p Params) Validate() error {
	if err := validateIsMainnet(p.IsMainnet); err != nil {
		return err
	}

	return nil
}

func validateIsMainnet(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
