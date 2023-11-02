package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// zero fee pool
func InitialFeePool() FeePool {
	return FeePool{
		CommunityPool:         sdk.DecCoins{},
		LiquidityProviderPool: sdk.DecCoins{},
	}
}

// ValidateGenesis validates the fee pool for a genesis state
func (f FeePool) ValidateGenesis() error {
	if f.CommunityPool.IsAnyNegative() {
		return fmt.Errorf("negative CommunityPool in distribution fee pool, is %v",
			f.CommunityPool)
	}
	if f.LiquidityProviderPool.IsAnyNegative() {
		return fmt.Errorf("negative LiquidityProviderPool in distribution fee pool, is %v",
			f.LiquidityProviderPool)
	}

	return nil
}
