package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// total accumulation tracker
type TotalAccum struct {
	UpdateHeight int64   `json:"update_height"`
	Accum        sdk.Dec `json:"accum"`
}

func NewTotalAccum(height int64) TotalAccum {
	return TotalAccum{
		UpdateHeight: height,
		Accum:        sdk.ZeroDec(),
	}
}

// update total validator accumulation factor for the new height
// CONTRACT: height should be greater than the old height
func (ta TotalAccum) UpdateForNewHeight(height int64, accumCreatedPerBlock sdk.Dec) TotalAccum {
	blocks := height - ta.UpdateHeight
	if blocks < 0 {
		panic("reverse updated for new height")
	}
	ta.Accum = ta.Accum.Add(accumCreatedPerBlock.MulInt(sdk.NewInt(blocks)))
	ta.UpdateHeight = height
	return ta
}

//___________________________________________________________________________________________

// global fee pool for distribution
type FeePool struct {
	TotalValAccum TotalAccum `json:"val_accum"`      // total valdator accum held by validators
	Pool          DecCoins   `json:"pool"`           // funds for all validators which have yet to be withdrawn
	CommunityPool DecCoins   `json:"community_pool"` // pool for community funds yet to be spent
}

// update total validator accumulation factor
// NOTE: Do not call this except from ValidatorDistInfo.TakeFeePoolRewards().
func (f FeePool) UpdateTotalValAccum(height int64, totalBondedTokens sdk.Dec) FeePool {
	f.TotalValAccum = f.TotalValAccum.UpdateForNewHeight(height, totalBondedTokens)
	return f
}

// zero fee pool
func InitialFeePool() FeePool {
	return FeePool{
		TotalValAccum: NewTotalAccum(0),
		Pool:          DecCoins{},
		CommunityPool: DecCoins{},
	}
}
