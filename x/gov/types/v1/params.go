package v1

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Default period for deposits & voting and min voting period
	DefaultVotingPeriod  time.Duration = time.Hour * 24 * 21 // 21 days
	DefaultDepositPeriod time.Duration = time.Hour * 24 * 14 // 14 days

	// MaxDecreaseSensitivityTargetDistanceDepositThrottler is the maximum
	// value that can be set for the sensitivity to target distance for
	// dynamic initial and total deposit (decreases). This value has been
	// empirically found to be sufficient for realistic usage. A higher
	// value would make the throttler too little sensitive to the distance
	// from the target when decreasing the deposit.
	MaxDecreaseSensitivityTargetDistanceDepositThrottler = 100

	// DefaultGovernorStatusChangePeriod is the default period that has to pass
	// before a governor can change their status (e.g. from active to inactive).
	DefaultGovernorStatusChangePeriod time.Duration = time.Hour * 24 * 28 // 28 days
)

// MinVotingPeriod is set in stone by the constitution at 21 days, but it can
// be overridden with ldflags for devnet/testnet environments (hence the use of
// the string type).
var MinVotingPeriod = "504h" // 21 days

func init() {
	// Ensure MinVotingPeriod can be parsed
	if _, err := time.ParseDuration(MinVotingPeriod); err != nil {
		panic(fmt.Sprintf("wrong value for MinVotingPeriod '%s': %v", MinVotingPeriod, err))
	}
}

// Default governance params
var (
	minVotingPeriod, _                    = time.ParseDuration(MinVotingPeriod)
	DefaultMinDepositTokens               = math.NewInt(10000000)
	DefaultMaxQuorum                      = math.LegacyNewDecWithPrec(50, 2)
	DefaultMinQuorum                      = math.LegacyNewDecWithPrec(10, 2)
	DefaultThreshold                      = math.LegacyNewDecWithPrec(667, 3)
	DefaultMaxConstitutionAmendmentQuorum = DefaultMaxQuorum
	DefaultMinConstitutionAmendmentQuorum = DefaultMinQuorum
	DefaultConstitutionAmendmentThreshold = math.LegacyNewDecWithPrec(9, 1)
	DefaultMaxLawQuorum                   = DefaultMaxQuorum
	DefaultMinLawQuorum                   = DefaultMinQuorum
	DefaultLawThreshold                   = math.LegacyNewDecWithPrec(9, 1)
	// DefaultMinInitialDepositRatio         = math.LegacyZeroDec()
	DefaultBurnProposalPrevote = false                           // set to false to replicate behavior of when this change was made (0.47)
	DefaultBurnVoteQuorom      = false                           // set to false to  replicate behavior of when this change was made (0.47)
	DefaultMinDepositRatio     = math.LegacyNewDecWithPrec(1, 2) // NOTE: backport from v50

	DefaultQuorumTimeout                                             = DefaultVotingPeriod - (time.Hour * 24 * 1) // disabled by default (DefaultQuorumCheckCount must be set to a non-zero value to enable)
	DefaultMaxVotingPeriodExtension                                  = DefaultVotingPeriod - DefaultQuorumTimeout // disabled by default (DefaultQuorumCheckCount must be set to a non-zero value to enable)
	DefaultQuorumCheckCount                                   uint64 = 0                                          //nolint:revive // disabled by default (0 means no check)
	DefaultMinDepositUpdatePeriod                                    = time.Hour * 24 * 7
	DefaultMinDepositDecreaseSensitivityTargetDistance        uint64 = 2
	DefaultMinDepositIncreaseRatio                                   = math.LegacyNewDecWithPrec(5, 2)
	DefaultMinDepositDecreaseRatio                                   = math.LegacyNewDecWithPrec(25, 3)
	DefaultTargetActiveProposals                              uint64 = 2
	DefaultMinInitialDepositFloorAmount                              = math.LegacyNewDecWithPrec(1, 2).MulInt(DefaultMinDepositTokens).TruncateInt()
	DefaultMinInitialDepositUpdatePeriod                             = time.Hour * 24
	DefaultMinInitialDepositDecreaseSensitivityTargetDistance uint64 = 2
	DefaultMinInitialDepositIncreaseRatio                            = math.LegacyNewDecWithPrec(1, 2)
	DefaultMinInitialDepositDecreaseRatio                            = math.LegacyNewDecWithPrec(5, 3)
	DefaultTargetProposalsInDepositPeriod                     uint64 = 5
	DefaultBurnDepositNoThreshold                                    = math.LegacyNewDecWithPrec(80, 2)
	DefaultProposalCancelRatio                                       = math.LegacyMustNewDecFromStr("0.5")
	DefaultProposalCancelDestAddress                                 = ""
	DefaultMaxGovernors                                       uint64 = 100
	DefaultMinGovernorSelfDelegation                                 = math.NewInt(1000_000000)
)

// Deprecated: NewDepositParams creates a new DepositParams object
func NewDepositParams(minDeposit sdk.Coins, maxDepositPeriod *time.Duration) DepositParams {
	return DepositParams{
		MinDeposit:       minDeposit,
		MaxDepositPeriod: maxDepositPeriod,
	}
}

// Deprecated: NewTallyParams creates a new TallyParams object
func NewTallyParams(quorum, threshold,
	constitutionAmendmentQuorum, constitutionAmendmentThreshold,
	lawQuorum, lawThreshold string,
) TallyParams {
	return TallyParams{
		Quorum:                         quorum,
		Threshold:                      threshold,
		ConstitutionAmendmentQuorum:    constitutionAmendmentQuorum,
		ConstitutionAmendmentThreshold: constitutionAmendmentThreshold,
		LawQuorum:                      lawQuorum,
		LawThreshold:                   lawThreshold,
	}
}

// Deprecated: NewVotingParams creates a new VotingParams object
func NewVotingParams(votingPeriod *time.Duration) VotingParams {
	return VotingParams{
		VotingPeriod: votingPeriod,
	}
}

// NewParams creates a new Params instance with given values.
func NewParams(
	// minDeposit sdk.Coins, // Deprecated in favor of dynamic min deposit
	maxDepositPeriod, votingPeriod time.Duration,
	threshold, constitutionAmendmentThreshold, lawThreshold string,
	// minInitialDepositRatio string, // Deprecated in favor of dynamic min initial deposit
	burnProposalDeposit, burnVoteQuorum bool, minDepositRatio string,
	quorumTimeout, maxVotingPeriodExtension time.Duration, quorumCheckCount uint64,
	minDepositFloor sdk.Coins, minDepositUpdatePeriod time.Duration, minDepositDecreaseSensitivityTargetDistance uint64,
	minDepositIncreaseRatio, minDepositDecreaseRatio string, targetActiveProposals uint64,
	minInitialDepositFloor sdk.Coins, minInitialDepositUpdatePeriod time.Duration, minInitialDepositDecreaseSensitivityTargetDistance uint64,
	minInitialDepositIncreaseRatio, minInitialDepositDecreaseRatio string, targetProposalsInDepositPeriod uint64,
	burnDepositNoThreshold string,
	maxQuorum, minQuorum string,
	maxConstitutionAmendmentQuorum, minConstitutionAmendmentQuorum string,
	maxLawQuorum, minLawQuorum string,
	proposalCancelRatio, proposalCancelDest string,
	governorStatusChangePeriod time.Duration, minGovernorSelfDelegation string,
) Params {
	return Params{
		// MinDeposit:                     minDeposit, // Deprecated in favor of dynamic min deposit
		MaxDepositPeriod:               &maxDepositPeriod,
		VotingPeriod:                   &votingPeriod,
		Threshold:                      threshold,
		ConstitutionAmendmentThreshold: constitutionAmendmentThreshold,
		LawThreshold:                   lawThreshold,
		// MinInitialDepositRatio:         minInitialDepositRatio, // Deprecated in favor of dynamic min deposit
		//
		ProposalCancelRatio:        proposalCancelRatio,
		ProposalCancelDest:         proposalCancelDest,
		BurnProposalDepositPrevote: burnProposalDeposit,
		BurnVoteQuorum:             burnVoteQuorum,
		MinDepositRatio:            minDepositRatio,
		QuorumTimeout:              &quorumTimeout,
		MaxVotingPeriodExtension:   &maxVotingPeriodExtension,
		QuorumCheckCount:           quorumCheckCount,
		MinDepositThrottler: &MinDepositThrottler{
			FloorValue:                        minDepositFloor,
			UpdatePeriod:                      &minDepositUpdatePeriod,
			DecreaseSensitivityTargetDistance: minDepositDecreaseSensitivityTargetDistance,
			IncreaseRatio:                     minDepositIncreaseRatio,
			DecreaseRatio:                     minDepositDecreaseRatio,
			TargetActiveProposals:             targetActiveProposals,
		},
		MinInitialDepositThrottler: &MinInitialDepositThrottler{
			FloorValue:                        minInitialDepositFloor,
			UpdatePeriod:                      &minInitialDepositUpdatePeriod,
			DecreaseSensitivityTargetDistance: minInitialDepositDecreaseSensitivityTargetDistance,
			IncreaseRatio:                     minInitialDepositIncreaseRatio,
			DecreaseRatio:                     minInitialDepositDecreaseRatio,
			TargetProposals:                   targetProposalsInDepositPeriod,
		},
		BurnDepositNoThreshold: burnDepositNoThreshold,
		QuorumRange: &QuorumRange{
			Max: maxQuorum,
			Min: minQuorum,
		},
		ConstitutionAmendmentQuorumRange: &QuorumRange{
			Max: maxConstitutionAmendmentQuorum,
			Min: minConstitutionAmendmentQuorum,
		},
		LawQuorumRange: &QuorumRange{
			Max: maxLawQuorum,
			Min: minLawQuorum,
		},
		GovernorStatusChangePeriod: &governorStatusChangePeriod,
		MinGovernorSelfDelegation:  minGovernorSelfDelegation,
	}
}

// DefaultParams returns the default governance params
func DefaultParams() Params {
	return NewParams(
		DefaultDepositPeriod,
		DefaultVotingPeriod,
		DefaultThreshold.String(),
		DefaultConstitutionAmendmentThreshold.String(),
		DefaultLawThreshold.String(),
		// DefaultMinInitialDepositRatio.String(),
		DefaultBurnProposalPrevote,
		DefaultBurnVoteQuorom,
		DefaultMinDepositRatio.String(),
		DefaultQuorumTimeout,
		DefaultMaxVotingPeriodExtension,
		DefaultQuorumCheckCount,
		GetDefaultMinDepositFloor(),
		DefaultMinDepositUpdatePeriod,
		DefaultMinDepositDecreaseSensitivityTargetDistance,
		DefaultMinDepositIncreaseRatio.String(),
		DefaultMinDepositDecreaseRatio.String(),
		DefaultTargetActiveProposals,
		GetDefaultMinInitialDepositFloor(),
		DefaultMinInitialDepositUpdatePeriod,
		DefaultMinInitialDepositDecreaseSensitivityTargetDistance,
		DefaultMinInitialDepositIncreaseRatio.String(),
		DefaultMinInitialDepositDecreaseRatio.String(),
		DefaultTargetProposalsInDepositPeriod,
		DefaultBurnDepositNoThreshold.String(),
		DefaultMaxQuorum.String(),
		DefaultMinQuorum.String(),
		DefaultMaxConstitutionAmendmentQuorum.String(),
		DefaultMinConstitutionAmendmentQuorum.String(),
		DefaultMaxLawQuorum.String(),
		DefaultMinLawQuorum.String(),
		DefaultProposalCancelRatio.String(),
		DefaultProposalCancelDestAddress,
		DefaultGovernorStatusChangePeriod,
		DefaultMinGovernorSelfDelegation.String(),
	)
}

// ValidateBasic performs basic validation on governance parameters.
func (p Params) ValidateBasic() error {
	// if minDeposit := sdk.Coins(p.MinDeposit); minDeposit.Empty() || !minDeposit.IsValid() {
	// 	return fmt.Errorf("invalid minimum deposit: %s", minDeposit)
	// }

	// if mindeposit is set, return error as it is deprecated
	// Q: is returning an error the best way to handle this? or perhaps just log a warning?
	//    after all this value is not used anymore in the codebase
	if len(p.MinDeposit) > 0 {
		return fmt.Errorf("manually setting min deposit is deprecated in favor of a dynamic min deposit")
	}
	if len(p.Quorum) > 0 {
		return fmt.Errorf("manually setting quorum is deprecated in favor of a dynamic quorum")
	}
	if len(p.ConstitutionAmendmentQuorum) > 0 {
		return fmt.Errorf("manually setting constitutionAmendmentQuorum is deprecated in favor of a dynamic quorum")
	}
	if len(p.LawQuorum) > 0 {
		return fmt.Errorf("manually setting lawQuorum is deprecated in favor of a dynamic quorum")
	}

	if p.MaxDepositPeriod == nil {
		return fmt.Errorf("maximum deposit period must not be nil: %d", p.MaxDepositPeriod)
	}

	if p.MaxDepositPeriod.Seconds() <= 0 {
		return fmt.Errorf("maximum deposit period must be positive: %d", p.MaxDepositPeriod)
	}

	for label, q := range map[string]string{
		"quorumRange.min":                      p.QuorumRange.Min,
		"quorumRange.max":                      p.QuorumRange.Max,
		"constitutionAmendmentQuorumRange.min": p.ConstitutionAmendmentQuorumRange.Min,
		"constitutionAmendmentQuorumRange.max": p.ConstitutionAmendmentQuorumRange.Max,
		"lawQuorumRange.min":                   p.LawQuorumRange.Min,
		"lawQuorumRange.max":                   p.LawQuorumRange.Max,
	} {
		quorum, err := math.LegacyNewDecFromStr(q)
		if err != nil {
			return fmt.Errorf("invalid %s string: %w", label, err)
		}
		if quorum.IsNegative() {
			return fmt.Errorf("%s must be positive: %s", label, quorum)
		}
		if quorum.GT(math.LegacyOneDec()) {
			return fmt.Errorf("%s too large: %s", label, quorum)
		}
	}

	if math.LegacyMustNewDecFromStr(p.QuorumRange.Max).LT(math.LegacyMustNewDecFromStr(p.QuorumRange.Min)) {
		return fmt.Errorf("quorum range max must be greater than or equal to min: %s", p.QuorumRange)
	}
	if math.LegacyMustNewDecFromStr(p.ConstitutionAmendmentQuorumRange.Max).LT(math.LegacyMustNewDecFromStr(p.ConstitutionAmendmentQuorumRange.Min)) {
		return fmt.Errorf("constitution amendment quorum range max must be greater than or equal to min: %s", p.ConstitutionAmendmentQuorumRange)
	}
	if math.LegacyMustNewDecFromStr(p.LawQuorumRange.Max).LT(math.LegacyMustNewDecFromStr(p.LawQuorumRange.Min)) {
		return fmt.Errorf("law quorum range max must be greater than or equal to min: %s", p.LawQuorumRange)
	}

	threshold, err := math.LegacyNewDecFromStr(p.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("vote threshold too large: %s", threshold)
	}

	amendmentThreshold, err := math.LegacyNewDecFromStr(p.ConstitutionAmendmentThreshold)
	if err != nil {
		return fmt.Errorf("invalid constitution amendment threshold string: %w", err)
	}
	if !amendmentThreshold.IsPositive() {
		return fmt.Errorf("constitution amendment threshold must be positive: %s", amendmentThreshold)
	}
	if amendmentThreshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("constitution amendment threshold too large: %s", amendmentThreshold)
	}
	if amendmentThreshold.LT(threshold) {
		return fmt.Errorf("constitution amendment threshold must be greater than or equal to governance threshold: %s", amendmentThreshold)
	}

	lawThreshold, err := math.LegacyNewDecFromStr(p.LawThreshold)
	if err != nil {
		return fmt.Errorf("invalid law threshold string: %w", err)
	}
	if !lawThreshold.IsPositive() {
		return fmt.Errorf("law threshold must be positive: %s", lawThreshold)
	}
	if lawThreshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("law threshold too large: %s", lawThreshold)
	}
	if lawThreshold.LT(threshold) {
		return fmt.Errorf("law threshold must be greater than or equal to governance threshold: %s", lawThreshold)
	}
	if lawThreshold.GT(amendmentThreshold) {
		return fmt.Errorf("law threshold must be less than or equal to constitution amendment threshold: %s", lawThreshold)
	}

	if p.VotingPeriod == nil {
		return fmt.Errorf("voting period must not be nil")
	}

	if p.VotingPeriod.Seconds() < minVotingPeriod.Seconds() {
		return fmt.Errorf("voting period must be at least %s: %s", minVotingPeriod.String(), p.VotingPeriod.String())
	}

	if len(p.MinInitialDepositRatio) > 0 {
		return fmt.Errorf("manually setting min initial deposit ratio is deprecated in favor of a dynamic min initial deposit")
	}

	minDepositRatio, err := math.LegacyNewDecFromStr(p.MinDepositRatio)
	if err != nil {
		return fmt.Errorf("invalid mininum deposit ratio of proposal: %w", err)
	}
	if minDepositRatio.IsNegative() {
		return fmt.Errorf("mininum deposit ratio of proposal must be positive: %s", minDepositRatio)
	}
	if minDepositRatio.GT(math.LegacyOneDec()) {
		return fmt.Errorf("mininum deposit ratio of proposal is too large: %s", minDepositRatio)
	}

	if p.QuorumCheckCount > 0 {
		// If quorum check is enabled, validate quorum check params
		if p.QuorumTimeout == nil {
			return fmt.Errorf("quorum timeout must not be nil")
		}
		if p.QuorumTimeout.Seconds() < 0 {
			return fmt.Errorf("quorum timeout must be 0 or greater: %s", p.QuorumTimeout)
		}
		if p.QuorumTimeout.Nanoseconds() >= p.VotingPeriod.Nanoseconds() {
			return fmt.Errorf("quorum timeout %s must be strictly less than the voting period %s", p.QuorumTimeout, p.VotingPeriod)
		}

		if p.MaxVotingPeriodExtension == nil {
			return fmt.Errorf("max voting period extension must not be nil")
		}
		if p.MaxVotingPeriodExtension.Nanoseconds() < p.VotingPeriod.Nanoseconds()-p.QuorumTimeout.Nanoseconds() {
			return fmt.Errorf("max voting period extension %s must be greater than or equal to the difference between the voting period %s and the quorum timeout %s", p.MaxVotingPeriodExtension, p.VotingPeriod, p.QuorumTimeout)
		}
	}

	if p.MinDepositThrottler == nil {
		return fmt.Errorf("min deposit throttler must not be nil")
	}

	if minDepositFloor := sdk.Coins(p.MinDepositThrottler.FloorValue); minDepositFloor.Empty() || !minDepositFloor.IsValid() {
		return fmt.Errorf("invalid minimum deposit floor: %s", minDepositFloor)
	}

	if p.MinDepositThrottler.UpdatePeriod == nil {
		return fmt.Errorf("minimum deposit update period must not be nil")
	}

	if p.MinDepositThrottler.UpdatePeriod.Seconds() <= 0 {
		return fmt.Errorf("minimum deposit update period must be positive: %s", p.MinDepositThrottler.UpdatePeriod)
	}

	if p.MinDepositThrottler.UpdatePeriod.Seconds() > p.VotingPeriod.Seconds() {
		return fmt.Errorf("minimum deposit update period must be less than or equal to the voting period: %s", p.MinDepositThrottler.UpdatePeriod)
	}

	if p.MinDepositThrottler.DecreaseSensitivityTargetDistance == 0 {
		return fmt.Errorf("minimum deposit sensitivity target distance must be positive: %d", p.MinDepositThrottler.DecreaseSensitivityTargetDistance)
	}

	if p.MinDepositThrottler.DecreaseSensitivityTargetDistance > MaxDecreaseSensitivityTargetDistanceDepositThrottler {
		return fmt.Errorf("minimum deposit sensitivity target distance must be less than or equal to %d: %d", MaxDecreaseSensitivityTargetDistanceDepositThrottler, p.MinDepositThrottler.DecreaseSensitivityTargetDistance)
	}

	minDepositIncreaseRatio, err := math.LegacyNewDecFromStr(p.MinDepositThrottler.IncreaseRatio)
	if err != nil {
		return fmt.Errorf("invalid minimum deposit increase ratio: %w", err)
	}
	if !minDepositIncreaseRatio.IsPositive() {
		return fmt.Errorf("minimum deposit increase ratio must be positive: %s", minDepositIncreaseRatio)
	}
	if minDepositIncreaseRatio.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("minimum deposit increase ratio too large: %s", minDepositIncreaseRatio)
	}

	minDepositDecreaseRatio, err := math.LegacyNewDecFromStr(p.MinDepositThrottler.DecreaseRatio)
	if err != nil {
		return fmt.Errorf("invalid minimum deposit decrease ratio: %w", err)
	}
	if !minDepositDecreaseRatio.IsPositive() {
		return fmt.Errorf("minimum deposit decrease ratio must be positive: %s", minDepositDecreaseRatio)
	}
	if minDepositDecreaseRatio.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("minimum deposit decrease ratio too large: %s", minDepositDecreaseRatio)
	}

	if p.MinInitialDepositThrottler == nil {
		return fmt.Errorf("min initial deposit throttler must not be nil")
	}

	if minInitialDepositFloor := sdk.Coins(p.MinInitialDepositThrottler.FloorValue); minInitialDepositFloor.Empty() || !minInitialDepositFloor.IsValid() {
		return fmt.Errorf("invalid minimum initial deposit floor: %s", minInitialDepositFloor)
	}

	if p.MinInitialDepositThrottler.UpdatePeriod == nil {
		return fmt.Errorf("minimum initial deposit update period must not be nil")
	}

	if p.MinInitialDepositThrottler.UpdatePeriod.Seconds() <= 0 {
		return fmt.Errorf("minimum initial deposit update period must be positive: %s", p.MinInitialDepositThrottler.UpdatePeriod)
	}

	if p.MinInitialDepositThrottler.UpdatePeriod.Seconds() > p.VotingPeriod.Seconds() {
		return fmt.Errorf("minimum initial deposit update period must be less than or equal to the voting period: %s", p.MinInitialDepositThrottler.UpdatePeriod)
	}

	if p.MinInitialDepositThrottler.DecreaseSensitivityTargetDistance == 0 {
		return fmt.Errorf("minimum initial deposit sensitivity target distance must be positive: %d", p.MinInitialDepositThrottler.DecreaseSensitivityTargetDistance)
	}

	if p.MinInitialDepositThrottler.DecreaseSensitivityTargetDistance > MaxDecreaseSensitivityTargetDistanceDepositThrottler {
		return fmt.Errorf("minimum initial deposit sensitivity target distance must be less than or equal to %d: %d", MaxDecreaseSensitivityTargetDistanceDepositThrottler, p.MinInitialDepositThrottler.DecreaseSensitivityTargetDistance)
	}

	minInitialDepositIncreaseRatio, err := math.LegacyNewDecFromStr(p.MinInitialDepositThrottler.IncreaseRatio)
	if err != nil {
		return fmt.Errorf("invalid minimum initial deposit increase ratio: %w", err)
	}

	if !minInitialDepositIncreaseRatio.IsPositive() {
		return fmt.Errorf("minimum initial deposit increase ratio must be positive: %s", minInitialDepositIncreaseRatio)
	}

	if minInitialDepositIncreaseRatio.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("minimum initial deposit increase ratio too large: %s", minInitialDepositIncreaseRatio)
	}

	minInitialDepositDecreaseRatio, err := math.LegacyNewDecFromStr(p.MinInitialDepositThrottler.DecreaseRatio)
	if err != nil {
		return fmt.Errorf("invalid minimum initial deposit decrease ratio: %w", err)
	}

	if !minInitialDepositDecreaseRatio.IsPositive() {
		return fmt.Errorf("minimum initial deposit decrease ratio must be positive: %s", minInitialDepositDecreaseRatio)
	}

	if minInitialDepositDecreaseRatio.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("minimum initial deposit decrease ratio too large: %s", minInitialDepositDecreaseRatio)
	}

	burnDepositNoThreshold, err := math.LegacyNewDecFromStr(p.BurnDepositNoThreshold)
	if err != nil {
		return fmt.Errorf("invalid burnDepositNoThreshold string: %w", err)
	}
	if burnDepositNoThreshold.LT(math.LegacyOneDec().Sub(amendmentThreshold)) {
		return fmt.Errorf("burnDepositNoThreshold cannot be lower than 1-amendmentThreshold")
	}
	if burnDepositNoThreshold.LT(math.LegacyOneDec().Sub(lawThreshold)) {
		return fmt.Errorf("burnDepositNoThreshold cannot be lower than 1-lawThreshold")
	}
	if burnDepositNoThreshold.LT(math.LegacyOneDec().Sub(threshold)) {
		return fmt.Errorf("burnDepositNoThreshold cannot be lower than 1-threshold")
	}
	if burnDepositNoThreshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("burnDepositNoThreshold too large: %s", burnDepositNoThreshold)
	}

	proposalCancelRate, err := math.LegacyNewDecFromStr(p.ProposalCancelRatio)
	if err != nil {
		return fmt.Errorf("invalid burn rate of cancel proposal: %w", err)
	}
	if proposalCancelRate.IsNegative() {
		return fmt.Errorf("burn rate of cancel proposal must be positive: %s", proposalCancelRate)
	}
	if proposalCancelRate.GT(math.LegacyOneDec()) {
		return fmt.Errorf("burn rate of cancel proposal is too large: %s", proposalCancelRate)
	}

	if len(p.ProposalCancelDest) != 0 {
		_, err := sdk.AccAddressFromBech32(p.ProposalCancelDest)
		if err != nil {
			return fmt.Errorf("deposits destination address is invalid: %s", p.ProposalCancelDest)
		}
	}

	if p.GovernorStatusChangePeriod == nil {
		return fmt.Errorf("governor status change period must not be nil: %d", p.GovernorStatusChangePeriod)
	}

	if p.GovernorStatusChangePeriod.Seconds() <= 0 {
		return fmt.Errorf("governor status change period must be positive: %d", p.GovernorStatusChangePeriod)
	}

	minGovernorSelfDelegation, ok := math.NewIntFromString(p.MinGovernorSelfDelegation)
	if !ok {
		return fmt.Errorf("invalid minimum governor self delegation: %s", p.MinGovernorSelfDelegation)
	}
	if minGovernorSelfDelegation.IsNegative() {
		return fmt.Errorf("minimum governor self delegation must be positive: %s", minGovernorSelfDelegation)
	}

	return nil
}

// GetDefaultMinDepositFloor returns the default minimum deposit floor
// required so the correct `sdk.DefaultBondDenom` is used.
func GetDefaultMinDepositFloor() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, DefaultMinDepositTokens))
}

// GetDefaultMinInitialDepositFloor returns the default minimum initial deposit floor
// required so the correct `sdk.DefaultBondDenom` is used.
func GetDefaultMinInitialDepositFloor() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, DefaultMinInitialDepositFloorAmount))
}
