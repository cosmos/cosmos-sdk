package v1

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default period for deposits & voting
const (
	DefaultPeriod                         time.Duration = time.Hour * 24 * 2 // 2 days
	DefaultExpeditedPeriod                time.Duration = time.Hour * 24 * 1 // 1 day
	DefaultMinExpeditedDepositTokensRatio               = 5
)

// Default governance params
var (
	DefaultMinDepositTokens          = sdk.NewInt(10000000)
	DefaultMinExpeditedDepositTokens = DefaultMinDepositTokens.Mul(math.NewInt(DefaultMinExpeditedDepositTokensRatio))
	DefaultQuorum                    = sdk.NewDecWithPrec(334, 3)
	DefaultThreshold                 = sdk.NewDecWithPrec(5, 1)
	DefaultExpeditedThreshold        = sdk.NewDecWithPrec(667, 3)
	DefaultVetoThreshold             = sdk.NewDecWithPrec(334, 3)
	DefaultMinInitialDepositRatio    = sdk.ZeroDec()
	DefaultProposalCancelRatio       = sdk.MustNewDecFromStr("0.5")
	DefaultProposalCancelDestAddress = ""
)

// Deprecated: NewDepositParams creates a new DepositParams object
func NewDepositParams(minDeposit sdk.Coins, maxDepositPeriod *time.Duration) DepositParams {
	return DepositParams{
		MinDeposit:       minDeposit,
		MaxDepositPeriod: maxDepositPeriod,
	}
}

// Deprecated: NewTallyParams creates a new TallyParams object
func NewTallyParams(quorum, threshold, vetoThreshold string) TallyParams {
	return TallyParams{
		Quorum:        quorum,
		Threshold:     threshold,
		VetoThreshold: vetoThreshold,
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
	minDeposit, expeditedminDeposit sdk.Coins, maxDepositPeriod, votingPeriod, expeditedVotingPeriod time.Duration,
	quorum, threshold, expeditedThreshold, vetoThreshold, minInitialDepositRatio, proposalCancelRatio, proposalCancelDest string,
) Params {
	return Params{
		MinDeposit:             minDeposit,
		ExpeditedMinDeposit:    expeditedminDeposit,
		MaxDepositPeriod:       &maxDepositPeriod,
		VotingPeriod:           &votingPeriod,
		ExpeditedVotingPeriod:  &expeditedVotingPeriod,
		Quorum:                 quorum,
		Threshold:              threshold,
		ExpeditedThreshold:     expeditedThreshold,
		VetoThreshold:          vetoThreshold,
		MinInitialDepositRatio: minInitialDepositRatio,
		ProposalCancelRatio:    proposalCancelRatio,
		ProposalCancelDest:     proposalCancelDest,
	}
}

// DefaultParams returns the default governance params
func DefaultParams() Params {
	return NewParams(
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, DefaultMinDepositTokens)),
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, DefaultMinExpeditedDepositTokens)),
		DefaultPeriod,
		DefaultPeriod,
		DefaultExpeditedPeriod,
		DefaultQuorum.String(),
		DefaultThreshold.String(),
		DefaultExpeditedThreshold.String(),
		DefaultVetoThreshold.String(),
		DefaultMinInitialDepositRatio.String(),
		DefaultProposalCancelRatio.String(),
		DefaultProposalCancelDestAddress,
	)
}

// ValidateBasic performs basic validation on governance parameters.
func (p Params) ValidateBasic() error {
	minDeposit := sdk.Coins(p.MinDeposit)
	if minDeposit.Empty() || !minDeposit.IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", minDeposit)
	}

	if minExpeditedDeposit := sdk.Coins(p.ExpeditedMinDeposit); minExpeditedDeposit.Empty() || !minExpeditedDeposit.IsValid() {
		return fmt.Errorf("invalid expedited minimum deposit: %s", minExpeditedDeposit)
	} else if minExpeditedDeposit.IsAllLTE(minDeposit) {
		return fmt.Errorf("expedited minimum deposit must be greater than minimum deposit: %s", minExpeditedDeposit)
	}

	if p.MaxDepositPeriod == nil {
		return fmt.Errorf("maximum deposit period must not be nil: %d", p.MaxDepositPeriod)
	}

	if p.MaxDepositPeriod.Seconds() <= 0 {
		return fmt.Errorf("maximum deposit period must be positive: %d", p.MaxDepositPeriod)
	}

	quorum, err := sdk.NewDecFromStr(p.Quorum)
	if err != nil {
		return fmt.Errorf("invalid quorum string: %w", err)
	}
	if quorum.IsNegative() {
		return fmt.Errorf("quorom cannot be negative: %s", quorum)
	}
	if quorum.GT(math.LegacyOneDec()) {
		return fmt.Errorf("quorom too large: %s", p.Quorum)
	}

	threshold, err := sdk.NewDecFromStr(p.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("vote threshold too large: %s", threshold)
	}

	expeditedThreshold, err := sdk.NewDecFromStr(p.ExpeditedThreshold)
	if err != nil {
		return fmt.Errorf("invalid expedited threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("expedited vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("expedited vote threshold too large: %s", threshold)
	}
	if expeditedThreshold.LTE(threshold) {
		return fmt.Errorf("expedited vote threshold %s, must be greater than the regular threshold %s", expeditedThreshold, threshold)
	}

	vetoThreshold, err := sdk.NewDecFromStr(p.VetoThreshold)
	if err != nil {
		return fmt.Errorf("invalid vetoThreshold string: %w", err)
	}
	if !vetoThreshold.IsPositive() {
		return fmt.Errorf("veto threshold must be positive: %s", vetoThreshold)
	}
	if vetoThreshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("veto threshold too large: %s", vetoThreshold)
	}

	if p.VotingPeriod == nil {
		return fmt.Errorf("voting period must not be nil: %d", p.VotingPeriod)
	}
	if p.VotingPeriod.Seconds() <= 0 {
		return fmt.Errorf("voting period must be positive: %s", p.VotingPeriod)
	}

	if p.ExpeditedVotingPeriod == nil {
		return fmt.Errorf("expedited voting period must not be nil: %d", p.ExpeditedVotingPeriod)
	}
	if p.ExpeditedVotingPeriod.Seconds() <= 0 {
		return fmt.Errorf("expedited voting period must be positive: %s", p.ExpeditedVotingPeriod)
	}
	if p.ExpeditedVotingPeriod.Seconds() >= p.VotingPeriod.Seconds() {
		return fmt.Errorf("expedited voting period %s must be strictly less that the regular voting period %s", p.ExpeditedVotingPeriod, p.VotingPeriod)
	}

	minInitialDepositRatio, err := sdk.NewDecFromStr(p.MinInitialDepositRatio)
	if err != nil {
		return fmt.Errorf("invalid mininum initial deposit ratio of proposal: %w", err)
	}
	if minInitialDepositRatio.IsNegative() {
		return fmt.Errorf("mininum initial deposit ratio of proposal must be positive: %s", minInitialDepositRatio)
	}
	if minInitialDepositRatio.GT(math.LegacyOneDec()) {
		return fmt.Errorf("mininum initial deposit ratio of proposal is too large: %s", minInitialDepositRatio)
	}

	proposalCancelRate, err := sdk.NewDecFromStr(p.ProposalCancelRatio)
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

	return nil
}
