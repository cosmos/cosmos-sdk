package types

import (
	fmt "fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func (r TokenizeShareRecord) GetModuleAddress() sdk.AccAddress {
	// NOTE: The module name is intentionally hard coded so that, if this
	// function were to move to a different module in future SDK version,
	// it would not break all the address lookups
	moduleName := "lsm"
	return address.Module(moduleName, []byte(r.ModuleAccount))
}

func (r TokenizeShareRecord) GetShareTokenDenom() string {
	return fmt.Sprintf("%s/%s", strings.ToLower(r.Validator), strconv.Itoa(int(r.Id)))
}
