package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// CheckExtensionOption accepts the `ExtensionOptionTieredTx` extension option.
func CheckExtensionOption(any *codectypes.Any) bool {
	_, ok := any.GetCachedValue().(*ExtensionOptionTieredTx)
	return ok
}
