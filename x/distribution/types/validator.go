package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// historical rewards for a validator
// height is implicit within the store key
type ValidatorHistoricalRewards struct {
	SeqValue       sdk.DecCoins `json:"seq_value"`
	ReferenceCount uint16       `json:"reference_count"`
}

// create a new ValidatorHistoricalRewards
func NewValidatorHistoricalRewards(seqValue sdk.DecCoins, referenceCount uint16) ValidatorHistoricalRewards {
	return ValidatorHistoricalRewards{
		SeqValue:       seqValue,
		ReferenceCount: referenceCount,
	}
}

// current rewards and current period for a validator
// kept as a running counter and incremented each block
// as long as the validator's tokens remain constant
type ValidatorCurrentRewards struct {
	Rewards sdk.DecCoins `json:"rewards"` // current rewards
	Period  uint64       `json:"period"`  // current period
}

// create a new ValidatorCurrentRewards
func NewValidatorCurrentRewards(rewards sdk.DecCoins, period uint64) ValidatorCurrentRewards {
	return ValidatorCurrentRewards{
		Rewards: rewards,
		Period:  period,
	}
}

// accumulated commission for a validator
// kept as a running counter, can be withdrawn at any time
type ValidatorAccumulatedCommission = sdk.DecCoins

// return the initial accumulated commission (zero)
func InitialValidatorAccumulatedCommission() ValidatorAccumulatedCommission {
	return ValidatorAccumulatedCommission{}
}

// validator slash event
// height is implicit within the store key
// needed to calculate appropriate amounts of staking token
// for delegations which withdraw after a slash has occurred
type ValidatorSlashEvent struct {
	ValidatorPeriod uint64  `json:"validator_period"` // period when the slash occurred
	Fraction        sdk.Dec `json:"fraction"`         // slash fraction
}

// create a new ValidatorSlashEvent
func NewValidatorSlashEvent(validatorPeriod uint64, fraction sdk.Dec) ValidatorSlashEvent {
	return ValidatorSlashEvent{
		ValidatorPeriod: validatorPeriod,
		Fraction:        fraction,
	}
}
