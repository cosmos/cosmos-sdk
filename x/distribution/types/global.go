package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// coins with decimal
type DecCoins []DecCoin

// Coins which can have additional decimal points
type DecCoin struct {
	Amount sdk.Dec
	Denom  string
}

//___________________________________________________________________________________________

// Global state for distribution
type Global struct {
	TotalValAccumUpdateHeight int64    // last height which the total validator accum was updated
	TotalValAccum             sdk.Dec  // total valdator accum held by validators
	Pool                      DecCoins // funds for all validators which have yet to be withdrawn
	CommunityPool             DecCoins // pool for community funds yet to be spent
}

// update total validator accumulation factor
func (g Global) UpdateTotalValAccum(height int64, totalbondedtokens dec) Global {
	blocks := height - g.totalvalaccumupdateheight
	g.totalvalaccum += totaldelshares * blocks
	g.totalvalaccumupdateheight = height
	return g
}
