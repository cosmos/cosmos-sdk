package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default parameter values
const (
	DefaultPubKeyChangeCost uint64 = 5000
)

// Parameter keys
var (
	KeyPubKeyChangeCost = []byte("PubKeyChangeCost")
)

var _ paramtypes.ParamSet = &Params{}

// NewParams creates a new Params object
func NewParams(
	pubKeyChangeCost uint64,
) Params {
	return Params{
		PubKeyChangeCost: pubKeyChangeCost,
	}
}

// ParamKeyTable for changepubkey module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of changepubkey module's parameters.
// nolint
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPubKeyChangeCost, &p.PubKeyChangeCost, validatePubKeyChangeCost),
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		PubKeyChangeCost: DefaultPubKeyChangeCost,
	}
}

func validatePubKeyChangeCost(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid pubkey change cost: %d", v)
	}

	return nil
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validatePubKeyChangeCost(p.PubKeyChangeCost); err != nil {
		return err
	}

	return nil
}
