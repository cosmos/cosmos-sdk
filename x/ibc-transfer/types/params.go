package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// DefaultTransfersEnabled enabled
	DefaultTransfersEnabled = true
)

// KeyTransfersEnabled is store's key for TransfersEnabled Params
var KeyTransfersEnabled = []byte("TransfersEnabled")

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for the ibc transfer module
func NewParams(enableTransfers bool) Params {
	return Params{
		TransfersEnabled: enableTransfers,
	}
}

// DefaultParams is the default parameter configuration for the ibc-transfer module
func DefaultParams() Params {
	return NewParams(DefaultTransfersEnabled)
}

// Validate all ibc-transfer module parameters
func (p Params) Validate() error {
	return validateTransfersEnabled(p.TransfersEnabled)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyTransfersEnabled, p.TransfersEnabled, validateTransfersEnabled),
	}
}

func validateTransfersEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
