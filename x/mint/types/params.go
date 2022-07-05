package types

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Parameter store keys
var (
	KeyMintDenom           = []byte("MintDenom")
	KeyInflationRateChange = []byte("InflationRateChange")
	KeyInflationMax        = []byte("InflationMax")
	KeyInflationMin        = []byte("InflationMin")
	KeyGoalBonded          = []byte("GoalBonded")
	KeyBlocksPerYear       = []byte("BlocksPerYear")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, inflationRateChange, inflationMax, inflationMin, goalBonded sdk.Dec, blocksPerYear uint64,
) Params {
	return Params{
		MintDenom:           mintDenom,
		InflationRateChange: inflationRateChange,
		InflationMax:        inflationMax,
		InflationMin:        inflationMin,
		GoalBonded:          goalBonded,
		BlocksPerYear:       blocksPerYear,
		MaxSupply:           maxSupply,
	}
}

// DefaultParams returns default x/mint module parameters.
func DefaultParams() Params {
	return Params{
		MintDenom:           sdk.DefaultBondDenom,
		InflationRateChange: math.LegacyNewDecWithPrec(13, 2),
		InflationMax:        math.LegacyNewDecWithPrec(5, 2),
		InflationMin:        math.LegacyNewDecWithPrec(0, 2),
		GoalBonded:          math.LegacyNewDecWithPrec(67, 2),
		BlocksPerYear:       uint64(60 * 60 * 8766 / 5), // assuming 5-second block times
		MaxSupply:           math.ZeroInt(),             // assuming zero is infinite
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
	if err := validateMaxSupply(p.MaxSupply); err != nil {
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

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyInflationRateChange, &p.InflationRateChange, validateInflationRateChange),
		paramtypes.NewParamSetPair(KeyInflationMax, &p.InflationMax, validateInflationMax),
		paramtypes.NewParamSetPair(KeyInflationMin, &p.InflationMin, validateInflationMin),
		paramtypes.NewParamSetPair(KeyGoalBonded, &p.GoalBonded, validateGoalBonded),
		paramtypes.NewParamSetPair(KeyBlocksPerYear, &p.BlocksPerYear, validateBlocksPerYear),
	}
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

func validateMaxSupply(v math.Int) error {
	if v.IsNegative() {
		return fmt.Errorf("max supply must be positive: %d", v)
	}

	return nil
}
