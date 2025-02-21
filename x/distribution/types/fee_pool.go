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

	if !f.CommunityPool.IsZero() {
		panic(fmt.Sprintf("CommunityPool must be zero in distribution fee pool as it should be specified in protocolpool, current value: %v", f.CommunityPool))
	}

	return nil
}
