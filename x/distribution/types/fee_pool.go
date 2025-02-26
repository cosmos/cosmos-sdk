package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitialFeePool initializes a zero fee pool
func InitialFeePool() FeePool {
	return FeePool{
		DecimalPool:   sdk.DecCoins{},
		CommunityPool: sdk.DecCoins{},
	}
}

// ValidateGenesis validates the fee pool for a genesis state
func (f FeePool) ValidateGenesis() error {
	if f.DecimalPool.IsAnyNegative() {
		return fmt.Errorf("negative DecimalPool in distribution fee pool, is %v", f.DecimalPool)
	}

	// Negative values in CommunityPool represent an invalid state that should never occur
	// We panic instead of returning an error to prevent the chain from starting with an invalid state
	if f.CommunityPool.IsAnyNegative() {
		panic(fmt.Sprintf("negative CommunityPool in distribution fee pool, is %v", f.CommunityPool))
	}

	return nil
}
