package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// historical rewards for a validator
// TODO add reference counter, ref https://github.com/cosmos/cosmos-sdk/pull/3099#discussion_r245747051
// height is implicit within the store key
type ValidatorHistoricalRewards = sdk.DecCoins

// current rewards for a validator
type ValidatorCurrentRewards struct {
	Rewards sdk.DecCoins `json:"rewards"`
	Period  uint64       `json:"period"`
}

// accumulated commission for a validator
type ValidatorAccumulatedCommission = sdk.DecCoins

// starting info for a delegator reward period
type DelegatorStartingInfo struct {
	PreviousPeriod uint64  `json:"previous_period"`
	Stake          sdk.Dec `json:"stake"`
	Height         uint64  `json:"height"`
}

// outstanding rewards for everyone
type OutstandingRewards = sdk.DecCoins

// validator slash event
type ValidatorSlashEvent struct {
	ValidatorPeriod uint64  `json:"validator_period"`
	Fraction        sdk.Dec `json:"fraction"`
}
