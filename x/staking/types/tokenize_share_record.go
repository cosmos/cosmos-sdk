package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (r TokenizeShareRecord) GetModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(r.ModuleAccount)
}
