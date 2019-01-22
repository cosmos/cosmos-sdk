package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// historical rewards for a validator
// height is implicit within the store key
// the reference count indicates the number of objects
// which might need to reference this historical entry
// at any point, it is equal to the number of outstanding
// delegations starting on the associated period, plus
// the number of slashes at the associated period, plus one
// per validator for the zeroeth period
type ValidatorHistoricalRewards struct {
	PeriodRewards  sdk.DecCoins `json:"period_rewards"`
	ReferenceCount uint16       `json:"reference_count"`
}

// create a new ValidatorHistoricalRewards
func NewValidatorHistoricalRewards(periodRewards sdk.DecCoins, referenceCount uint16) ValidatorHistoricalRewards {
	return ValidatorHistoricalRewards{
		PeriodRewards:  periodRewards,
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

func (vs ValidatorSlashEvent) String() string {
	return fmt.Sprintf(`Period:   %d
Fraction: %s`, vs.ValidatorPeriod, vs.Fraction)
}

// ValidatorSlashEvents is a collection of ValidatorSlashEvent
type ValidatorSlashEvents []ValidatorSlashEvent

func (vs ValidatorSlashEvents) String() string {
	out := "Validator Slash Events:\n"
	for i, sl := range vs {
		out += fmt.Sprintf(`  Slash %d:
    Period:   %d
    Fraction: %s
`, i, sl.ValidatorPeriod, sl.Fraction)
	}
	return strings.TrimSpace(out)
}
