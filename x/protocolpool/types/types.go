package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// DistributionInfo represents the information about distributed funds for a recipient.
type DistributionInfo struct {
	Address sdk.AccAddress `json:"address"`
	Amount  sdk.Coin       `json:"amount"`
}
