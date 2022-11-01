package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (r TokenizeShareRecord) GetModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(r.ModuleAccount)
}

func (r TokenizeShareRecord) GetShareTokenDenom() string {
	return fmt.Sprintf("%s/%d", strings.ToLower(r.Validator), r.Id)
}
