package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// starting info for a delegator reward period
// tracks the previous validator period, the delegation's amount
// of staking token, and the creation height (to check later on
// if any slashes have occurred)
// NOTE that even though validators are slashed to whole staking tokens, the
// delegators within the validator may be left with less than a full token,
// thus sdk.Dec is used
type DelegatorStartingInfo struct {
	PreviousPeriod uint64  `json:"previous_period"` // period at which the delegation should withdraw starting from
	Stake          sdk.Dec `json:"stake"`           // amount of staking token delegated
	Height         uint64  `json:"creation_height"` // height at which delegation was created
}

// create a new DelegatorStartingInfo
func NewDelegatorStartingInfo(previousPeriod uint64, stake sdk.Dec, height uint64) DelegatorStartingInfo {
	return DelegatorStartingInfo{
		PreviousPeriod: previousPeriod,
		Stake:          stake,
		Height:         height,
	}
}
