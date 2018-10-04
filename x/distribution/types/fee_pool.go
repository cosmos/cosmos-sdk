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

// update total validator accumulation factor
func (ta TotalAccum) Update(height int64, accumCreatedPerBlock sdk.Dec) TotalAccum {
	blocks := height - ta.UpdateHeight
	ta.Accum = ta.Accum.Add(accumCreatedPerBlock.Mul(sdk.NewDec(blocks)))
	ta.UpdateHeight = height
	return ta
}

//___________________________________________________________________________________________

// global fee pool for distribution
type FeePool struct {
	ValAccum      TotalAccum `json:"val_accum"`      // total valdator accum held by validators
	Pool          DecCoins   `json:"pool"`           // funds for all validators which have yet to be withdrawn
	CommunityPool DecCoins   `json:"community_pool"` // pool for community funds yet to be spent
}

// update total validator accumulation factor
func (f FeePool) UpdateTotalValAccum(height int64, totalBondedTokens sdk.Dec) FeePool {
	f.ValAccum = f.ValAccum.Update(height, totalBondedTokens)
	return f
}

// zero fee pool
func InitialFeePool() FeePool {
	return FeePool{
		ValAccum:      NewTotalAccum(0),
		Pool:          DecCoins{},
		CommunityPool: DecCoins{},
	}
}
