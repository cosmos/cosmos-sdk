package bank

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	// ModuleName is the name of the module
	ModuleName = "bank"
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName
	// DefaultSendEnabled enabled
	DefaultSendEnabled = true
)

// ParamStoreKeySendEnabled is store's key for SendEnabled
var ParamStoreKeySendEnabled = []byte("sendenabled")

// ParamKeyTable type declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeySendEnabled, false,
	)
}
