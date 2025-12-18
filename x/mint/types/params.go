package types

import (
	"errors"
	"fmt"
	"math"
	"strings"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams returns Params instance with the given values.
func NewParams(mintDenom string, inflationRateChange, inflationMax, inflationMin, goalBonded sdkmath.LegacyDec, blocksPerYear uint64) Params {
	return Params{
		MintDenom:           mintDenom,
		InflationRateChange: inflationRateChange,
		InflationMax:        inflationMax,
		InflationMin:        inflationMin,
		GoalBonded:          goalBonded,
		BlocksPerYear:       blocksPerYear,
	}
}

// DefaultParams returns default x/mint module parameters.
func DefaultParams() Params {
	return Params{
		MintDenom:           sdk.DefaultBondDenom,
		InflationRateChange: sdkmath.LegacyNewDecWithPrec(13, 2),
		InflationMax:        sdkmath.LegacyNewDecWithPrec(20, 2),
		InflationMin:        sdkmath.LegacyNewDecWithPrec(7, 2),
		GoalBonded:          sdkmath.LegacyNewDecWithPrec(67, 2),
		BlocksPerYear:       uint64(60 * 60 * 8766 / 5), // assuming 5 second block times
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

	return nil
}

func validateMintDenom(denom string) error {
	if strings.TrimSpace(denom) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(denom); err != nil {
		return err
	}

	return nil
}

func validateInflationRateChange(rateChange sdkmath.LegacyDec) error {
	if rateChange.IsNil() {
		return fmt.Errorf("inflation rate change cannot be nil: %s", rateChange)
	}
	if rateChange.IsNegative() {
		return fmt.Errorf("inflation rate change cannot be negative: %s", rateChange)
	}
	if rateChange.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("inflation rate change too large: %s", rateChange)
	}

	return nil
}

func validateInflationMax(inflationMax sdkmath.LegacyDec) error {
	if inflationMax.IsNil() {
		return fmt.Errorf("max inflation cannot be nil: %s", inflationMax)
	}
	if inflationMax.IsNegative() {
		return fmt.Errorf("max inflation cannot be negative: %s", inflationMax)
	}
	if inflationMax.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("max inflation too large: %s", inflationMax)
	}

	return nil
}

func validateInflationMin(inflationMin sdkmath.LegacyDec) error {
	if inflationMin.IsNil() {
		return fmt.Errorf("min inflation cannot be nil: %s", inflationMin)
	}
	if inflationMin.IsNegative() {
		return fmt.Errorf("min inflation cannot be negative: %s", inflationMin)
	}
	if inflationMin.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("min inflation too large: %s", inflationMin)
	}

	return nil
}

func validateGoalBonded(goalBonded sdkmath.LegacyDec) error {
	if goalBonded.IsNil() {
		return fmt.Errorf("goal bonded cannot be nil: %s", goalBonded)
	}
	if goalBonded.IsNegative() || goalBonded.IsZero() {
		return fmt.Errorf("goal bonded must be positive: %s", goalBonded)
	}
	if goalBonded.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("goal bonded too large: %s", goalBonded)
	}

	return nil
}

func validateBlocksPerYear(blocksPerYear uint64) error {
	if blocksPerYear == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", blocksPerYear)
	}

	if blocksPerYear > math.MaxInt64 {
		return fmt.Errorf("blocks per year too large: %d, maximum value is: %d", blocksPerYear, math.MaxInt64)
	}

	return nil
}
