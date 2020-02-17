package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName
	// DefaultSendEnabled enabled
	DefaultSendEnabled = true
)

// ParamStoreKeySendEnabled is store's key for SendEnabled
var ParamStoreKeySendEnabled = []byte("sendenabled")

// ParamKeyTable type declaration for parameters
func ParamKeyTable() types.KeyTable {
	return types.NewKeyTable(
		types.NewParamSetPair(ParamStoreKeySendEnabled, false, validateSendEnabled),
	)
}

func validateSendEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
