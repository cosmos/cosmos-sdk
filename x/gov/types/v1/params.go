package v1

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default period for deposits & voting
const (
	DefaultPeriod time.Duration = time.Hour * 24 * 2 // 2 days
)

// Default governance params
var (
	DefaultMinDepositTokens = sdk.NewInt(10000000)
	DefaultQuorum           = sdk.NewDecWithPrec(334, 3)
	DefaultThreshold        = sdk.NewDecWithPrec(5, 1)
	DefaultVetoThreshold    = sdk.NewDecWithPrec(334, 3)
)

// // Equal checks equality of DepositParams
// func (dp DepositParams) Equal(dp2 DepositParams) bool {
// 	return sdk.Coins(dp.MinDeposit).IsEqual(dp2.MinDeposit) && dp.MaxDepositPeriod == dp2.MaxDepositPeriod
// }

func validateDepositParams(i interface{}) error {
	v, ok := i.(DepositParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !sdk.Coins(v.MinDeposit).IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", v.MinDeposit)
	}
	if v.MaxDepositPeriod == nil || v.MaxDepositPeriod.Seconds() <= 0 {
		return fmt.Errorf("maximum deposit period must be positive: %d", v.MaxDepositPeriod)
	}

	return nil
}

// // Equal checks equality of TallyParams
// func (tp TallyParams) Equal(other TallyParams) bool {
// 	return tp.Quorum == other.Quorum && tp.Threshold == other.Threshold && tp.VetoThreshold == other.VetoThreshold
// }

func validateTallyParams(i interface{}) error {
	v, ok := i.(TallyParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	quorum, err := sdk.NewDecFromStr(v.Quorum)
	if err != nil {
		return fmt.Errorf("invalid quorum string: %w", err)
	}
	if quorum.IsNegative() {
		return fmt.Errorf("quorom cannot be negative: %s", quorum)
	}
	if quorum.GT(sdk.OneDec()) {
		return fmt.Errorf("quorom too large: %s", v)
	}

	threshold, err := sdk.NewDecFromStr(v.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(sdk.OneDec()) {
		return fmt.Errorf("vote threshold too large: %s", v)
	}

	vetoThreshold, err := sdk.NewDecFromStr(v.VetoThreshold)
	if err != nil {
		return fmt.Errorf("invalid vetoThreshold string: %w", err)
	}
	if !vetoThreshold.IsPositive() {
		return fmt.Errorf("veto threshold must be positive: %s", vetoThreshold)
	}
	if vetoThreshold.GT(sdk.OneDec()) {
		return fmt.Errorf("veto threshold too large: %s", v)
	}

	return nil
}

// // Equal checks equality of TallyParams
// func (vp VotingParams) Equal(other VotingParams) bool {
// 	return vp.VotingPeriod == other.VotingPeriod
// }

func validateVotingParams(i interface{}) error {
	v, ok := i.(VotingParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.VotingPeriod == nil {
		return errors.New("voting period must not be nil")
	}

	if v.VotingPeriod.Seconds() <= 0 {
		return fmt.Errorf("voting period must be positive: %s", v.VotingPeriod)
	}

	return nil
}

func NewParams(
	minDeposit sdk.Coins, maxDepositPeriod time.Duration, votingPeriod time.Duration,
	quorum string, threshold string, vetoThreshold string,
) Params {
	return Params{
		MinDeposit:       minDeposit,
		MaxDepositPeriod: &maxDepositPeriod,
		VotingPeriod:     &votingPeriod,
		Quorum:           quorum,
		Threshold:        threshold,
		VetoThreshold:    vetoThreshold,
	}
}

// DefaultParams default governance params
func DefaultParams() Params {
	return NewParams(
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, DefaultMinDepositTokens)),
		DefaultPeriod,
		DefaultPeriod,
		DefaultQuorum.String(),
		DefaultThreshold.String(),
		DefaultVetoThreshold.String(),
	)
}

func (p Params) ValidateBasic() error {

	if !sdk.Coins(p.MinDeposit).IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", p.MinDeposit)
	}
	if p.MaxDepositPeriod == nil || p.MaxDepositPeriod.Seconds() <= 0 {
		return fmt.Errorf("maximum deposit period must be positive: %d", p.MaxDepositPeriod)
	}

	quorum, err := sdk.NewDecFromStr(p.Quorum)
	if err != nil {
		return fmt.Errorf("invalid quorum string: %w", err)
	}
	if quorum.IsNegative() {
		return fmt.Errorf("quorom cannot be negative: %s", quorum)
	}
	if quorum.GT(sdk.OneDec()) {
		return fmt.Errorf("quorom too large: %s", p.Quorum)
	}

	threshold, err := sdk.NewDecFromStr(p.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(sdk.OneDec()) {
		return fmt.Errorf("vote threshold too large: %s", threshold)
	}

	vetoThreshold, err := sdk.NewDecFromStr(p.VetoThreshold)
	if err != nil {
		return fmt.Errorf("invalid vetoThreshold string: %w", err)
	}
	if !vetoThreshold.IsPositive() {
		return fmt.Errorf("veto threshold must be positive: %s", vetoThreshold)
	}
	if vetoThreshold.GT(sdk.OneDec()) {
		return fmt.Errorf("veto threshold too large: %s", vetoThreshold)
	}

	if p.VotingPeriod == nil {
		return errors.New("voting period must not be nil")
	}

	if p.VotingPeriod.Seconds() <= 0 {
		return fmt.Errorf("voting period must be positive: %s", p.VotingPeriod)
	}

	return nil
}
