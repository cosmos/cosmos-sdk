package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ParamStoreKeyMinGasPrices store key
var ParamStoreKeyMinGasPrices = []byte("MetadataParam")

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		Twitter:       "",
		Telegram:      "",
		Discord:       "",
		Github:        "",
		Website:       "",
		CoingeckoId:   "",
		CoinImageLink: "",
		Constitution:  "",
		Other:         []*ChainSpecific{},
	}
}

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ValidateBasic performs basic validation.
func (p Params) ValidateBasic() error {
	return nil
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			ParamStoreKeyMinGasPrices, &p, validateParams,
		),
	}
}

func validateParams(i interface{}) error {
	return nil
}

// Validate checks that no strings are set too long or too short

func (params Params) Validate() error {
	for _, p := range params.Other {
		if p.Id == "" {
			return fmt.Errorf("empty id for value %s", p.Value)
		}
	}

	return nil
}
