package funds

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// MigrateFunds migrates the distribution module funds to pool module
func MigrateFunds(ctx sdk.Context, bankKeeper types.BankKeeper, feePool types.FeePool, macc, poolMacc sdk.ModuleAccountI) error {
	poolBal := sdk.NormalizeCoins(feePool.CommunityPool)
	distrbalances := bankKeeper.GetAllBalances(ctx, macc.GetAddress())
	if distrbalances.IsZero() || distrbalances.IsAllLT(poolBal) {
		return fmt.Errorf("%s module account balance is less than FeePool balance", macc.GetName())
	}

	// transfer feepool funds from the distribution module account to pool module account
	err := bankKeeper.SendCoinsFromModuleToModule(ctx, macc.GetName(), poolMacc.GetName(), poolBal)
	if err != nil {
		return err
	}

	// check the migrated balance from pool module account is same as fee pool balance
	balances := bankKeeper.GetAllBalances(ctx, poolMacc.GetAddress())
	if !balances.Equal(poolBal) {
		return fmt.Errorf("pool module account balance is not same as FeePool balance after migration")
	}
	return nil
}
