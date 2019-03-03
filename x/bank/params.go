package bank

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = "bank"
)

// ParamKeyTable type declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable()
}
