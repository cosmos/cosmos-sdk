package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// coins with decimal
type DecCoins []DecCoin

// Coins which can have additional decimal points
type DecCoin struct {
	Amount sdk.Dec `json:"amount"`
	Denom  string  `json:"denom"`
}

// total accumulation tracker
type TotalAccum struct {
	UpdateHeight int64   `json:"update_height"`
	Accum        sdk.Dec `json:"accum"`
}

// update total validator accumulation factor
func (ta TotalAccum) Update(height int64, accumCreatedPerBlock sdk.Dec) TotalAccum {
	blocks := height - ta.UpdateHeight
	f.Accum += accumCreatedPerBlock.Mul(sdk.NewDec(blocks))
	ta.UpdateHeight = height
	return ta
}

//___________________________________________________________________________________________

// global fee pool for distribution
type FeePool struct {
	ValAccum      TotalAccum `json:"val_accum"`      // total valdator accum held by validators
	Pool          DecCoins   `json:"pool"`           // funds for all validators which have yet to be withdrawn
	CommunityPool DecCoins   `json:"community_pool"` // pool for community funds yet to be spent}
}

// update total validator accumulation factor
func (f FeePool) UpdateTotalValAccum(height int64, totalBondedTokens Dec) FeePool {
	f.ValAccum = f.ValAccum.Update(height, totalBondedTokens)
	return f
}
