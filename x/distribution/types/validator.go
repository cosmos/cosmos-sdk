package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// historical rewards for a validator
// height is implicit within the store key
// cumulative reward ratio is the sum from the zeroeth period
// until this period of rewards / tokens, per the spec
// The reference count indicates the number of objects
// which might need to reference this historical entry
// at any point.
// ReferenceCount =
//    number of outstanding delegations which ended the associated period (and might need to read that record)
//  + number of slashes which ended the associated period (and might need to read that record)
//  + one per validator for the zeroeth period, set on initialization
type ValidatorHistoricalRewards struct {
	CumulativeRewardRatio sdk.DecCoins `json:"cumulative_reward_ratio"`
	ReferenceCount        uint16       `json:"reference_count"`
}

// create a new ValidatorHistoricalRewards
func NewValidatorHistoricalRewards(cumulativeRewardRatio sdk.DecCoins, referenceCount uint16) ValidatorHistoricalRewards {
	return ValidatorHistoricalRewards{
		CumulativeRewardRatio: cumulativeRewardRatio,
		ReferenceCount:        referenceCount,
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
	TokensBurn      sdk.Int `json:"tokens_burned"`    // tokens burned
}

// create a new ValidatorSlashEvent
func NewValidatorSlashEvent(validatorPeriod uint64, tokensBurn sdk.Int) ValidatorSlashEvent {
	return ValidatorSlashEvent{
		ValidatorPeriod: validatorPeriod,
		TokensBurn:      tokensBurn,
	}
}

func (vs ValidatorSlashEvent) String() string {
	return fmt.Sprintf(`Period:   %d
Tokens Burn: %s`, vs.ValidatorPeriod, vs.TokensBurn)
}

// ValidatorSlashEvents is a collection of ValidatorSlashEvent
type ValidatorSlashEvents []ValidatorSlashEvent

func (vs ValidatorSlashEvents) String() string {
	out := "Validator Slash Events:\n"
	for i, sl := range vs {
		out += fmt.Sprintf(`  Slash %d:
    Period:   %d
    Tokens Burn: %s
`, i, sl.ValidatorPeriod, sl.TokensBurn)
	}
	return strings.TrimSpace(out)
}

// outstanding (un-withdrawn) rewards for a validator
// inexpensive to track, allows simple sanity checks
type ValidatorOutstandingRewards = sdk.DecCoins
