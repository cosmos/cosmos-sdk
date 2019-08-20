package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// global fee pool for distribution
type FeePool struct {
	CommunityPool sdk.DecCoins `json:"community_pool" yaml:"community_pool"` // pool for community funds yet to be spent
}

// zero fee pool
func InitialFeePool() FeePool {
	return FeePool{
		CommunityPool: sdk.DecCoins{},
	}
}

// ValidateGenesis validates the fee pool for a genesis state
func (f FeePool) ValidateGenesis() error {
	if f.CommunityPool.IsAnyNegative() {
		return fmt.Errorf("negative CommunityPool in distribution fee pool, is %v",
			f.CommunityPool)
	}

	return nil
}
