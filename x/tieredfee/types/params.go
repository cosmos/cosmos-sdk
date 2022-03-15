package types

import (
	"errors"

	"sigs.k8s.io/yaml"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// KeyTiers is store's key for Tiers Params
	KeyTiers               = []byte("tiers")
	DefaultInitialGasPrice = sdk.NewDecCoins(sdk.NewDecCoinFromDec("atom", sdk.NewDecWithPrec(1, 6)))
)

// ParamKeyTable for tieredfee module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams is the default parameter configuration for the tieredfee module
func DefaultParams() Params {
	return Params{
		Tiers: []TierParams{
			{
				Priority:          10,
				InitialGasPrice:   DefaultInitialGasPrice,
				ParentGasTarget:   10000000,
				ChangeDenominator: 8,
				MinGasPrice:       nil,
				MaxGasPrice:       nil,
			},
		},
	}
}

// Validate all tieredfee module parameters
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyTiers, &p.Tiers, validateTiers),
	}
}

func validateTiers(value interface{}) error {
	// TODO
	_, ok := value.([]TierParams)
	if !ok {
		return errors.New("invalid type")
	}
	return nil
}
