package v1

import (
	"fmt"
	"time"

	"cosmossdk.io/core/address"
	sdkmath "cosmossdk.io/math"

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
	DefaultMinDepositTokens             = sdkmath.NewInt(10000000)
	DefaultMinExpeditedDepositTokens    = DefaultMinDepositTokens.Mul(sdkmath.NewInt(DefaultMinExpeditedDepositTokensRatio))
	DefaultQuorum                       = sdkmath.LegacyNewDecWithPrec(334, 3)
	DefaultYesQuorum                    = sdkmath.LegacyNewDecWithPrec(0, 1)
	DefaultExpeditedQuorum              = sdkmath.LegacyNewDecWithPrec(500, 3)
	DefaultThreshold                    = sdkmath.LegacyNewDecWithPrec(5, 1)
	DefaultExpeditedThreshold           = sdkmath.LegacyNewDecWithPrec(667, 3)
	DefaultVetoThreshold                = sdkmath.LegacyNewDecWithPrec(334, 3)
	DefaultMinInitialDepositRatio       = sdkmath.LegacyZeroDec()
	DefaultProposalCancelRatio          = sdkmath.LegacyMustNewDecFromStr("0.5")
	DefaultProposalCancelDestAddress    = ""
	DefaultProposalCancelMaxPeriod      = sdkmath.LegacyMustNewDecFromStr("0.5")
	DefaultBurnProposalPrevote          = false // set to false to replicate behavior of when this change was made (0.47)
	DefaultBurnVoteQuorum               = false // set to false to  replicate behavior of when this change was made (0.47)
	DefaultBurnVoteVeto                 = true  // set to true to replicate behavior of when this change was made (0.47)
	DefaultMinDepositRatio              = sdkmath.LegacyMustNewDecFromStr("0.01")
	DefaultOptimisticRejectedThreshold  = sdkmath.LegacyMustNewDecFromStr("0.1")
	DefaultOptimisticAuthorizedAddreses = []string(nil)
)

// NewParams creates a new Params instance with given values.
func NewParams(
	minDeposit, expeditedminDeposit sdk.Coins,
	maxDepositPeriod, votingPeriod, expeditedVotingPeriod time.Duration,
	quorum, yesQuorum, expeditedQuorum, threshold, expeditedThreshold, vetoThreshold, minInitialDepositRatio, proposalCancelRatio, proposalCancelDest, proposalMaxCancelVotingPeriod string,
	burnProposalDeposit, burnVoteQuorum, burnVoteVeto bool,
	minDepositRatio, optimisticRejectedThreshold string,
	optimisticAuthorizedAddresses []string,
) Params {
	return Params{
		MinDeposit:                    minDeposit,
		ExpeditedMinDeposit:           expeditedminDeposit,
		MaxDepositPeriod:              &maxDepositPeriod,
		VotingPeriod:                  &votingPeriod,
		ExpeditedVotingPeriod:         &expeditedVotingPeriod,
		Quorum:                        quorum,
		YesQuorum:                     yesQuorum,
		ExpeditedQuorum:               expeditedQuorum,
		Threshold:                     threshold,
		ExpeditedThreshold:            expeditedThreshold,
		VetoThreshold:                 vetoThreshold,
		MinInitialDepositRatio:        minInitialDepositRatio,
		ProposalCancelRatio:           proposalCancelRatio,
		ProposalCancelDest:            proposalCancelDest,
		ProposalCancelMaxPeriod:       proposalMaxCancelVotingPeriod,
		BurnProposalDepositPrevote:    burnProposalDeposit,
		BurnVoteQuorum:                burnVoteQuorum,
		BurnVoteVeto:                  burnVoteVeto,
		MinDepositRatio:               minDepositRatio,
		OptimisticRejectedThreshold:   optimisticRejectedThreshold,
		OptimisticAuthorizedAddresses: optimisticAuthorizedAddresses,
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
		DefaultYesQuorum.String(),
		DefaultExpeditedQuorum.String(),
		DefaultThreshold.String(),
		DefaultExpeditedThreshold.String(),
		DefaultVetoThreshold.String(),
		DefaultMinInitialDepositRatio.String(),
		DefaultProposalCancelRatio.String(),
		DefaultProposalCancelDestAddress,
		DefaultProposalCancelMaxPeriod.String(),
		DefaultBurnProposalPrevote,
		DefaultBurnVoteQuorum,
		DefaultBurnVoteVeto,
		DefaultMinDepositRatio.String(),
		DefaultOptimisticRejectedThreshold.String(),
		DefaultOptimisticAuthorizedAddreses,
	)
}

// ValidateBasic performs basic validation on governance parameters.
func (p Params) ValidateBasic(addressCodec address.Codec) error {
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

	quorum, err := sdkmath.LegacyNewDecFromStr(p.Quorum)
	if err != nil {
		return fmt.Errorf("invalid quorum string: %w", err)
	}
	if quorum.IsNegative() {
		return fmt.Errorf("quorum cannot be negative: %s", quorum)
	}
	if quorum.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("quorum too large: %s", p.Quorum)
	}

	yesQuorum, err := sdkmath.LegacyNewDecFromStr(p.YesQuorum)
	if err != nil {
		return fmt.Errorf("invalid yes_quorum string: %w", err)
	}
	if yesQuorum.IsNegative() {
		return fmt.Errorf("yes_quorum cannot be negative: %s", yesQuorum)
	}
	if yesQuorum.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("yes_quorum too large: %s", p.YesQuorum)
	}

	expeditedQuorum, err := sdkmath.LegacyNewDecFromStr(p.ExpeditedQuorum)
	if err != nil {
		return fmt.Errorf("invalid expedited_quorum string: %w", err)
	}
	if expeditedQuorum.IsNegative() {
		return fmt.Errorf("expedited_quorum cannot be negative: %s", expeditedQuorum)
	}
	if expeditedQuorum.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("expedited_quorum too large: %s", p.ExpeditedQuorum)
	}

	threshold, err := sdkmath.LegacyNewDecFromStr(p.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("vote threshold too large: %s", threshold)
	}

	expeditedThreshold, err := sdkmath.LegacyNewDecFromStr(p.ExpeditedThreshold)
	if err != nil {
		return fmt.Errorf("invalid expedited threshold string: %w", err)
	}
	if !expeditedThreshold.IsPositive() {
		return fmt.Errorf("expedited vote threshold must be positive: %s", expeditedThreshold)
	}
	if expeditedThreshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("expedited vote threshold too large: %s", expeditedThreshold)
	}
	if expeditedThreshold.LTE(threshold) {
		return fmt.Errorf("expedited vote threshold %s, must be greater than the regular threshold %s", expeditedThreshold, threshold)
	}

	vetoThreshold, err := sdkmath.LegacyNewDecFromStr(p.VetoThreshold)
	if err != nil {
		return fmt.Errorf("invalid vetoThreshold string: %w", err)
	}
	if !vetoThreshold.IsPositive() {
		return fmt.Errorf("veto threshold must be positive: %s", vetoThreshold)
	}
	if vetoThreshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("veto threshold too large: %s", vetoThreshold)
	}

	optimisticRejectedThreshold, err := sdkmath.LegacyNewDecFromStr(p.OptimisticRejectedThreshold)
	if err != nil {
		return fmt.Errorf("invalid optimistic rejected threshold string: %w", err)
	}

	if !optimisticRejectedThreshold.IsPositive() {
		return fmt.Errorf("optimistic rejected threshold must be positive: %s", optimisticRejectedThreshold)
	}

	if optimisticRejectedThreshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("optimistic rejected threshold too large: %s", optimisticRejectedThreshold)
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
		return fmt.Errorf("expedited voting period %s must be strictly less than the regular voting period %s", p.ExpeditedVotingPeriod, p.VotingPeriod)
	}

	for _, addr := range p.OptimisticAuthorizedAddresses {
		if _, err := addressCodec.StringToBytes(addr); err != nil {
			return fmt.Errorf("invalid optimistic authorized address: %s", addr)
		}
	}

	minInitialDepositRatio, err := sdkmath.LegacyNewDecFromStr(p.MinInitialDepositRatio)
	if err != nil {
		return fmt.Errorf("invalid minimum initial deposit ratio of proposal: %w", err)
	}
	if minInitialDepositRatio.IsNegative() {
		return fmt.Errorf("minimum initial deposit ratio of proposal must be positive: %s", minInitialDepositRatio)
	}
	if minInitialDepositRatio.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("minimum initial deposit ratio of proposal is too large: %s", minInitialDepositRatio)
	}

	proposalCancelRate, err := sdkmath.LegacyNewDecFromStr(p.ProposalCancelRatio)
	if err != nil {
		return fmt.Errorf("invalid burn rate of cancel proposal: %w", err)
	}
	if proposalCancelRate.IsNegative() {
		return fmt.Errorf("burn rate of cancel proposal must be positive: %s", proposalCancelRate)
	}
	if proposalCancelRate.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("burn rate of cancel proposal is too large: %s", proposalCancelRate)
	}

	proposalCancelMaxPeriod, err := sdkmath.LegacyNewDecFromStr(p.ProposalCancelMaxPeriod)
	if err != nil {
		return fmt.Errorf("invalid max cancel period of cancel proposal: %w", err)
	}
	if proposalCancelMaxPeriod.IsNegative() {
		return fmt.Errorf("max cancel period of cancel proposal must be positive: %s", proposalCancelMaxPeriod)
	}
	if proposalCancelMaxPeriod.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("max cancel period of cancel proposal is too large: %s", proposalCancelMaxPeriod)
	}

	if len(p.ProposalCancelDest) != 0 {
		_, err := addressCodec.StringToBytes(p.ProposalCancelDest)
		if err != nil {
			return fmt.Errorf("deposits destination address is invalid: %s", p.ProposalCancelDest)
		}
	}

	return nil
}

// ValidateBasic performs basic validation on governance parameters.
func (p MessageBasedParams) ValidateBasic() error {
	if p.VotingPeriod == nil {
		return fmt.Errorf("voting period must not be nil: %d", p.VotingPeriod)
	}
	if p.VotingPeriod.Seconds() <= 0 {
		return fmt.Errorf("voting period must be positive: %s", p.VotingPeriod)
	}

	quorum, err := sdkmath.LegacyNewDecFromStr(p.Quorum)
	if err != nil {
		return fmt.Errorf("invalid quorum string: %w", err)
	}
	if quorum.IsNegative() {
		return fmt.Errorf("quorum cannot be negative: %s", quorum)
	}
	if quorum.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("quorum too large: %s", p.Quorum)
	}

	yesQuorum, err := sdkmath.LegacyNewDecFromStr(p.YesQuorum)
	if err != nil {
		return fmt.Errorf("invalid yes_quorum string: %w", err)
	}
	if yesQuorum.IsNegative() {
		return fmt.Errorf("yes_quorum cannot be negative: %s", yesQuorum)
	}
	if yesQuorum.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("yes_quorum too large: %s", p.YesQuorum)
	}

	vetoThreshold, err := sdkmath.LegacyNewDecFromStr(p.VetoThreshold)
	if err != nil {
		return fmt.Errorf("invalid vetoThreshold string: %w", err)
	}
	if !vetoThreshold.IsPositive() {
		return fmt.Errorf("veto threshold must be positive: %s", vetoThreshold)
	}
	if vetoThreshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("veto threshold too large: %s", vetoThreshold)
	}

	threshold, err := sdkmath.LegacyNewDecFromStr(p.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("vote threshold too large: %s", threshold)
	}

	return nil
}
