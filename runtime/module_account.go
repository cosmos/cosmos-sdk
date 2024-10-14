package runtime

import (
	"cosmossdk.io/core/address"

	addresstypes "github.com/cosmos/cosmos-sdk/types/address"
)

var _ address.ModuleAccount = (*ModuleAccountService)(nil)

type ModuleAccountService struct{}

func (h ModuleAccountService) GetModuleAccountAddress(name string) []byte {
	return addresstypes.Module(name)
}
