package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// historical rewards for a validator
type ValidatorHistoricalRewards struct {
	AccumulatedFees sdk.DecCoins `json:"accumulated_fees"`
	TotalPower      sdk.Dec      `json:"total_power"`
}

// current rewards for a validator
type ValidatorCurrentRewards struct {
	Rewards sdk.DecCoins `json:"rewards"`
	Period  uint64       `json:"period"`
}

// accumulated commission for a validator
type ValidatorAccumulatedCommission = sdk.DecCoins

// starting period for a delegator's rewards
type DelegatorStartingPeriod = uint64

// outstanding rewards for everyone
type OutstandingRewards = sdk.DecCoins
