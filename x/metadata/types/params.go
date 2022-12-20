package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ParamStoreKeyMetadataPrices store key
var ParamStoreKeyMetadataPrices = []byte("MetadataParam")

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
		paramtypes.NewParamSetPair(ParamStoreKeyMetadataPrices, &p, validateParams),
	}
}

func validateParams(i interface{}) error {
	return nil
}

// Validate checks that no strings are set too long or too short

func (params Params) Validate() error {

	if len(params.Twitter) > 15 {
		return fmt.Errorf("Twitter handle too long")
	}

	// check that links start with http
	for _, link := range []string{params.Telegram, params.Discord, params.Github, params.Website} {
		if len(link) > 0 && (link[:4] != "http" || link[:3] != "www") {
			return fmt.Errorf("link %s does not start with http(s) or www", link)
		}
	}

	for _, p := range params.Other {
		if p.Id == "" {
			return fmt.Errorf("empty id for value %s", p.Value)
		}
	}

	return nil
}
