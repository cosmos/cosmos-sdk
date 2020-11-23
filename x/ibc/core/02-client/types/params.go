package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// DefaultAllowedClients are "06-solomachine" and "07-tendermint"
	DefaultAllowedClients = []string{exported.Solomachine, exported.Tendermint}
	// DefaultHistoricalEntries is 100.
	DefaultHistoricalEntries uint32 = 100

	// KeyAllowedClients is the store key for AllowedClients Params
	KeyAllowedClients = []byte("AllowedClients")
	// KeyHistoricalEntries is the store key for HistoricalEntries Params
	KeyHistoricalEntries = []byte("HistoricalEntries")
)

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for the ibc client.
func NewParams(historicalEntries uint32, allowedClients ...string) Params {
	return Params{
		AllowedClients:    allowedClients,
		HistoricalEntries: historicalEntries,
	}
}

// DefaultParams is the default parameter configuration for the ibc client.
func DefaultParams() Params {
	return NewParams(DefaultHistoricalEntries, DefaultAllowedClients...)
}

// Validate all ibc client submodule parameters
func (p Params) Validate() error {
	if err := validateClients(p.AllowedClients); err != nil {
		return err
	}

	return validateHistoricalEntries(p.HistoricalEntries)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAllowedClients, p.AllowedClients, validateClients),
		paramtypes.NewParamSetPair(KeyHistoricalEntries, &p.HistoricalEntries, validateHistoricalEntries),
	}
}

// IsAllowedClient checks if the given client type is registered on the allowlist.
func (p Params) IsAllowedClient(clientType string) bool {
	for _, allowedClient := range p.AllowedClients {
		if allowedClient == clientType {
			return true
		}
	}
	return false
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

func validateHistoricalEntries(i interface{}) error {
	entries, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid historical entries parameter type: %T", i)
	}

	if entries == 0 {
		return errors.New("historical entries parameter cannot be 0")
	}

	return nil
}
