package v4

import (
	"context"
	"fmt"

	"cosmossdk.io/core/moduleaccounts"
	"cosmossdk.io/x/distribution/types"
)

// MigrateFunds migrates the distribution module funds to pool module
func MigrateFunds(ctx context.Context,
	bankKeeper types.BankKeeper,
	feePool types.FeePool,
	modAccSvc moduleaccounts.Service,
	distrModule, poolModule string,
) (types.FeePool, error) {
	poolBal, remainder := feePool.CommunityPool.TruncateDecimal()

	macc := modAccSvc.Address(distrModule)
	poolMacc := modAccSvc.Address(poolModule)

	distrbalances := bankKeeper.GetAllBalances(ctx, macc)
	if distrbalances.IsZero() || distrbalances.IsAllLT(poolBal) {
		return types.FeePool{}, fmt.Errorf("%s module account balance is less than FeePool balance", distrModule)
	}

	// transfer feepool funds from the distribution module account to pool module account
	if err := bankKeeper.SendCoinsFromModuleToModule(ctx, distrModule, poolModule, poolBal); err != nil {
		return types.FeePool{}, err
	}

	// check the migrated balance from pool module account is same as fee pool balance
	balances := bankKeeper.GetAllBalances(ctx, poolMacc)
	if !balances.Equal(poolBal) {
		return types.FeePool{}, fmt.Errorf("pool module account balance is not same as FeePool balance after migration, %s != %s", balances, poolBal)
	}

	return types.FeePool{DecimalPool: remainder}, nil
}
