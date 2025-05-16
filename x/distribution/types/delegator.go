package types

import sdkmath "github.com/cosmos/cosmos-sdk/math/v2"

// NewDelegatorStartingInfo creates a new DelegatorStartingInfo
func NewDelegatorStartingInfo(previousPeriod uint64, stake sdkmath.LegacyDec, height uint64) DelegatorStartingInfo {
	return DelegatorStartingInfo{
		PreviousPeriod: previousPeriod,
		Stake:          stake,
		Height:         height,
	}
}
