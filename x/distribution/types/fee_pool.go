package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tendermint/libs/common"
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

// update total accumulation factor for the new height
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

// get total accumulation factor for the given height
// CONTRACT: height should be greater than the old height
func (ta TotalAccum) GetAccum(height int64, accumCreatedPerBlock sdk.Dec) sdk.Dec {
	blocks := height - ta.UpdateHeight
	if blocks < 0 {
		panic("reverse updated for new height")
	}
	return ta.Accum.Add(accumCreatedPerBlock.MulInt(sdk.NewInt(blocks)))
}

// update total validator accumulation factor for the new height
// CONTRACT: height should be greater than the old height
func (ta TotalAccum) UpdateForNewHeightDEBUG(height int64, accumCreatedPerBlock sdk.Dec) TotalAccum {
	blocks := height - ta.UpdateHeight
	if blocks < 0 {
		panic("reverse updated for new height")
	}
	if !accumCreatedPerBlock.IsZero() && blocks != 0 {
		fmt.Println(
			cmn.Blue(
				fmt.Sprintf("FP Add %v * %v = %v, + %v (old) => %v (new)",
					accumCreatedPerBlock.String(), sdk.NewInt(blocks),
					accumCreatedPerBlock.MulInt(sdk.NewInt(blocks)).String(),
					ta.Accum.String(),
					ta.Accum.Add(accumCreatedPerBlock.MulInt(sdk.NewInt(blocks))).String(),
				),
			),
		)
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
	f.TotalValAccum = f.TotalValAccum.UpdateForNewHeightDEBUG(height, totalBondedTokens)
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
		Pool:          DecCoins{},
		CommunityPool: DecCoins{},
	}
}
