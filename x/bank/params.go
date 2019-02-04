package bank

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	// default paramspace for params keeper
	DefaultParamspace = "bank"
	// default send enabled
	DefaultSendEnabled = true
)

// ParamStoreKeySendEnabled is store's key for SendEnabled
var ParamStoreKeySendEnabled = []byte("sendenabled")

// type declaration for parameters
func ParamTypeTable() params.TypeTable {
	return params.NewTypeTable(
		ParamStoreKeySendEnabled, false,
	)
}
