package keeper

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName
	// DefaultSendEnabled enabled
	DefaultSendEnabled = true // NOTE: this must always be consistent with the bank's param
)

// ParamStoreKeySendEnabled is store's key for SendEnabled
var ParamStoreKeySendEnabled = []byte("sendenabled")

// ParamKeyTable type declaration for parameters for the supply module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeySendEnabled, false,
	)
}
