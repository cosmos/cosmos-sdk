package types

import (
	"fmt"
	"strings"

	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// DefaultAllowedClients are "Solo Machine" and "Tendermint"
	DefaultAllowedClients = []string{solomachinetypes.SoloMachine, ibctmtypes.Tendermint}

	// KeyAllowedClients is store's key for AllowedClients Params
	KeyAllowedClients = []byte("AllowedClients")
)

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for the ibc transfer module
func NewParams(allowedClients ...string) Params {
	return Params{
		AllowedClients: allowedClients,
	}
}

// DefaultParams is the default parameter configuration for the ibc-transfer module
func DefaultParams() Params {
	return NewParams(DefaultAllowedClients...)
}

// Validate all ibc-transfer module parameters
func (p Params) Validate() error {
	return validateClients(p.AllowedClients)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAllowedClients, p.AllowedClients, validateClients),
	}
}

func validateClients(i interface{}) error {
	clients, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for i, clientType := range clients {
		if strings.TrimSpace(clientType) == "" {
			return fmt.Errorf("client type %d cannot be blank", i)
		}
	}

	return nil
}
