package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected auth Account Keeper (noalias)
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}
