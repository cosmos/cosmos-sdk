package types

import (
	"errors"
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewParams(
	mintDenom string,
	inflationRateChange,
	inflationMax,
	inflationMin,
	goalBonded sdk.Dec,
	blocksPerYear uint64,
	maxMintableAmount uint64,
	mintedAmountPerBlock sdk.Dec,
	yearlyReduction sdk.Dec,
) Params {

	return Params{
		MintDenom:            mintDenom,
		InflationRateChange:  inflationRateChange,
		InflationMax:         inflationMax,
		InflationMin:         inflationMin,
		GoalBonded:           goalBonded,
		BlocksPerYear:        blocksPerYear,
		MaxMintableAmount:    maxMintableAmount,
		MintedAmountPerBlock: mintedAmountPerBlock,
		YearlyReduction:      yearlyReduction,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:            sdk.DefaultBondDenom,
		InflationRateChange:  sdk.NewDecWithPrec(13, 2),
		InflationMax:         sdk.NewDecWithPrec(20, 2),
		InflationMin:         sdk.NewDecWithPrec(7, 2),
		GoalBonded:           sdk.NewDecWithPrec(67, 2),
		BlocksPerYear:        uint64(60 * 60 * 8766 / 5), // assuming 5 second block times
		MaxMintableAmount:    uint64(1000000000000000),
		MintedAmountPerBlock: sdk.NewDecWithPrec(18*1e6, 0),
		YearlyReduction:      sdk.NewDecWithPrec(1262079466, 10),
	}
}

// validate params
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
	if err := validateMaxMintableAmount(p.MaxMintableAmount); err != nil {
		return err
	}
	if err := validateYearlyReduction(p.YearlyReduction); err != nil {
		return err
	}
	if err := validateMintedAmountPerBlock(p.MintedAmountPerBlock); err != nil {
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

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateInflationRateChange(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("inflation rate change cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("inflation rate change too large: %s", v)
	}

	return nil
}

func validateInflationMax(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("max inflation too large: %s", v)
	}

	return nil
}

func validateInflationMin(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("min inflation too large: %s", v)
	}

	return nil
}

func validateGoalBonded(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("goal bonded cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("goal bonded too large: %s", v)
	}

	return nil
}

func validateBlocksPerYear(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}

func validateMintedAmountPerBlock(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("minted amount per block cannot be negative: %s", v)
	}

	return nil
}

func validateYearlyReduction(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("goal bonded cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("yearly reduction too large (limited to 100%): %s", v)
	}

	return nil
}

func validateMaxMintableAmount(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max mintable amount must be positive: %d", v)
	}

	return nil
}
