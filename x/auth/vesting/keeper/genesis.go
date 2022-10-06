package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
)

// InitGenesis initializes the vesting module state based on genesis state.
func (vk VestingKeeper) InitGenesis(ctx sdk.Context) {
	vk.accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) bool {
		if va, ok := account.(exported.VestingAccount); ok {
			vk.AddVestingAccount(ctx, va.GetAddress())
		}
		return false
	})
}
