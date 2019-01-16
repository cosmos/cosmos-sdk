package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// global fee pool for distribution
type FeePool struct {
	TotalValAccum TotalAccum `json:"val_accum"`      // total valdator accum held by validators
	ValPool       DecCoins   `json:"val_pool"`       // funds for all validators which have yet to be withdrawn
	CommunityPool DecCoins   `json:"community_pool"` // pool for community funds yet to be spent
}

// update total validator accumulation factor
// NOTE: Do not call this except from ValidatorDistInfo.TakeFeePoolRewards().
func (f FeePool) UpdateTotalValAccum(height int64, totalBondedTokens sdk.Dec) FeePool {
	f.TotalValAccum = f.TotalValAccum.UpdateForNewHeight(height, totalBondedTokens)
	return f
}

// get the total validator accum for the fee pool without modifying the state
func (f FeePool) GetTotalValAccum(height int64, totalBondedTokens sdk.Dec) sdk.Dec {
	return f.TotalValAccum.GetAccum(height, totalBondedTokens)
}

// zero fee pool
func InitialFeePool() FeePool {
	return FeePool{
		TotalValAccum: NewTotalAccum(0),
		ValPool:       DecCoins{},
		CommunityPool: DecCoins{},
	}
}

// ValidateGenesis validates the fee pool for a genesis state
func (f FeePool) ValidateGenesis() error {
	if f.TotalValAccum.Accum.IsNegative() {
		return fmt.Errorf("negative accum in distribution fee pool, is %v",
			f.TotalValAccum.Accum.String())
	}
	if f.TotalValAccum.UpdateHeight < 0 {
		return fmt.Errorf("negative update height in distribution fee pool, is %v",
			f.TotalValAccum.UpdateHeight)
	}
	if f.ValPool.HasNegative() {
		return fmt.Errorf("negative ValPool in distribution fee pool, is %v",
			f.ValPool)
	}
	if f.CommunityPool.HasNegative() {
		return fmt.Errorf("negative CommunityPool in distribution fee pool, is %v",
			f.CommunityPool)
	}

	return nil
}
