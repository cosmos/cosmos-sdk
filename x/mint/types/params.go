package types

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/math"
	epochtypes "cosmossdk.io/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams returns Params instance with the given values.
func NewParams(mintDenom string, inflationRateChange, inflationMax, inflationMin, goalBonded math.LegacyDec, blocksPerYear uint64, epochIdentifier string, reductionPeriodInEpochs int64, reductionFactor math.LegacyDec, mintingRewardsDistrStartEpoch int64, epochProvisions math.LegacyDec) Params {
	return Params{
		MintDenom:                            mintDenom,
		InflationRateChange:                  inflationRateChange,
		InflationMax:                         inflationMax,
		InflationMin:                         inflationMin,
		GoalBonded:                           goalBonded,
		BlocksPerYear:                        blocksPerYear,
		EpochIdentifier:                      epochIdentifier,
		ReductionPeriodInEpochs:              reductionPeriodInEpochs,
		ReductionFactor:                      reductionFactor,
		MintingRewardsDistributionStartEpoch: mintingRewardsDistrStartEpoch,
		GenesisEpochProvisions:               epochProvisions,
	}
}

// DefaultParams returns default x/mint module parameters.
func DefaultParams() Params {
	return Params{
		MintDenom:                            sdk.DefaultBondDenom,
		InflationRateChange:                  math.LegacyNewDecWithPrec(13, 2),
		InflationMax:                         math.LegacyNewDecWithPrec(20, 2),
		InflationMin:                         math.LegacyNewDecWithPrec(7, 2),
		GoalBonded:                           math.LegacyNewDecWithPrec(67, 2),
		BlocksPerYear:                        uint64(60 * 60 * 8766 / 5),
		EpochIdentifier:                      "week",                          // 1 week
		ReductionPeriodInEpochs:              156,                             // 3 years
		ReductionFactor:                      math.LegacyNewDecWithPrec(5, 1), // 0.5
		MintingRewardsDistributionStartEpoch: 0,
		GenesisEpochProvisions:               math.LegacyNewDec(5000000),
	}
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateInflationRateChange(p.InflationRateChange); err != nil {
		return err
	}
	if err := validateInflationMax(p.InflationMax); err != nil {
		return err
	}
	if err := validateInflationMin(p.InflationMin); err != nil {
		return err
	}
	if err := validateGoalBonded(p.GoalBonded); err != nil {
		return err
	}
	if err := validateBlocksPerYear(p.BlocksPerYear); err != nil {
		return err
	}
	if p.InflationMax.LT(p.InflationMin) {
		return fmt.Errorf(
			"max inflation (%s) must be greater than or equal to min inflation (%s)",
			p.InflationMax, p.InflationMin,
		)
	}
	if err := validateGenesisEpochProvisions(p.GenesisEpochProvisions); err != nil {
		return err
	}
	if err := epochtypes.ValidateEpochIdentifierInterface(p.EpochIdentifier); err != nil {
		return err
	}
	if err := validateReductionPeriodInEpochs(p.ReductionPeriodInEpochs); err != nil {
		return err
	}
	if err := validateReductionFactor(p.ReductionFactor); err != nil {
		return err
	}
	if err := validateMintingRewardsDistributionStartEpoch(p.MintingRewardsDistributionStartEpoch); err != nil {
		return err
	}

	return nil
}

func validateMintDenom(v string) error {
	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateInflationRateChange(v math.LegacyDec) error {
	if v.IsNil() {
		return fmt.Errorf("inflation rate change cannot be nil: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("inflation rate change cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("inflation rate change too large: %s", v)
	}

	return nil
}

func validateInflationMax(v math.LegacyDec) error {
	if v.IsNil() {
		return fmt.Errorf("max inflation cannot be nil: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("max inflation cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("max inflation too large: %s", v)
	}

	return nil
}

func validateInflationMin(v math.LegacyDec) error {
	if v.IsNil() {
		return fmt.Errorf("min inflation cannot be nil: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("min inflation cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("min inflation too large: %s", v)
	}

	return nil
}

func validateGoalBonded(v math.LegacyDec) error {
	if v.IsNil() {
		return fmt.Errorf("goal bonded cannot be nil: %s", v)
	}
	if v.IsNegative() || v.IsZero() {
		return fmt.Errorf("goal bonded must be positive: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("goal bonded too large: %s", v)
	}

	return nil
}

func validateBlocksPerYear(v uint64) error {
	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}

func validateGenesisEpochProvisions(v math.LegacyDec) error {
	if v.IsNegative() {
		return fmt.Errorf("genesis epoch provision must be non-negative")
	}

	return nil
}

func validateReductionPeriodInEpochs(v int64) error {
	if v <= 0 {
		return fmt.Errorf("max validators must be positive: %d", v)
	}

	return nil
}

func validateReductionFactor(v math.LegacyDec) error {
	if v.GT(math.LegacyNewDec(1)) {
		return fmt.Errorf("reduction factor cannot be greater than 1")
	}

	if v.IsNegative() {
		return fmt.Errorf("reduction factor cannot be negative")
	}

	return nil
}

func validateMintingRewardsDistributionStartEpoch(v int64) error {
	if v < 0 {
		return fmt.Errorf("start epoch must be non-negative")
	}

	return nil
}
