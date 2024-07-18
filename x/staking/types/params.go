package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Staking params default values
const (
	// DefaultUnbondingTime reflects three weeks in seconds as the default
	// unbonding time.
	// TODO: Justify our choice of default here.
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 3

	// Default maximum number of bonded validators
	DefaultMaxValidators uint32 = 100

	// Default maximum entries in a UBD/RED pair
	DefaultMaxEntries uint32 = 7

	// DefaultHistorical entries is 10000. Apps that don't use IBC can ignore this
	// value by not adding the staking module to the application module manager's
	// SetOrderBeginBlockers.
	DefaultHistoricalEntries uint32 = 10000
)

var (
	// DefaultMinCommissionRate is set to 0%
	DefaultMinCommissionRate = math.LegacyZeroDec()

	// ValidatorBondFactor of -1 indicates that it's disabled
	ValidatorBondCapDisabled = math.LegacyNewDecFromInt(math.NewInt(-1))

	// DefaultValidatorBondFactor is set to -1 (disabled)
	DefaultValidatorBondFactor = ValidatorBondCapDisabled
	// DefaultGlobalLiquidStakingCap is set to 100%
	DefaultGlobalLiquidStakingCap = math.LegacyOneDec()
	// DefaultValidatorLiquidStakingCap is set to 100%
	DefaultValidatorLiquidStakingCap = math.LegacyOneDec()
)

// NewParams creates a new Params instance
func NewParams(unbondingTime time.Duration,
	maxValidators,
	maxEntries,
	historicalEntries uint32,
	bondDenom string,
	minCommissionRate math.LegacyDec,
	validatorBondFactor math.LegacyDec,
	globalLiquidStakingCap math.LegacyDec,
	validatorLiquidStakingCap math.LegacyDec,
) Params {
	return Params{
		UnbondingTime:     unbondingTime,
		MaxValidators:     maxValidators,
		MaxEntries:        maxEntries,
		HistoricalEntries: historicalEntries,

		BondDenom:                 bondDenom,
		MinCommissionRate:         minCommissionRate,
		ValidatorBondFactor:       validatorBondFactor,
		GlobalLiquidStakingCap:    globalLiquidStakingCap,
		ValidatorLiquidStakingCap: validatorLiquidStakingCap,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultUnbondingTime,
		DefaultMaxValidators,
		DefaultMaxEntries,
		DefaultHistoricalEntries,
		sdk.DefaultBondDenom,
		DefaultMinCommissionRate,
		DefaultValidatorBondFactor,
		DefaultGlobalLiquidStakingCap,
		DefaultValidatorLiquidStakingCap,
	)
}

// unmarshal the current staking params value from store key or panic
func MustUnmarshalParams(cdc *codec.LegacyAmino, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}

	return params
}

// unmarshal the current staking params value from store key
func UnmarshalParams(cdc *codec.LegacyAmino, value []byte) (params Params, err error) {
	err = cdc.Unmarshal(value, &params)
	if err != nil {
		return
	}

	return
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateUnbondingTime(p.UnbondingTime); err != nil {
		return err
	}

	if err := validateMaxValidators(p.MaxValidators); err != nil {
		return err
	}

	if err := validateMaxEntries(p.MaxEntries); err != nil {
		return err
	}

	if err := validateBondDenom(p.BondDenom); err != nil {
		return err
	}

	if err := validateMinCommissionRate(p.MinCommissionRate); err != nil {
		return err
	}

	if err := validateHistoricalEntries(p.HistoricalEntries); err != nil {
		return err
	}

	if err := validateValidatorBondFactor(p.ValidatorBondFactor); err != nil {
		return err
	}

	if err := validateGlobalLiquidStakingCap(p.GlobalLiquidStakingCap); err != nil {
		return err
	}

	if err := validateValidatorLiquidStakingCap(p.ValidatorLiquidStakingCap); err != nil {
		return err
	}

	return nil
}

func validateUnbondingTime(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("unbonding time must be positive: %d", v)
	}

	return nil
}

func validateMaxValidators(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max validators must be positive: %d", v)
	}

	return nil
}

func validateMaxEntries(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max entries must be positive: %d", v)
	}

	return nil
}

func validateHistoricalEntries(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBondDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("bond denom cannot be blank")
	}

	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func ValidatePowerReduction(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(math.NewInt(1)) {
		return fmt.Errorf("power reduction cannot be lower than 1")
	}

	return nil
}

func validateMinCommissionRate(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("minimum commission rate cannot be nil: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("minimum commission rate cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("minimum commission rate cannot be greater than 100%%: %s", v)
	}

	return nil
}

func validateValidatorBondFactor(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() && !v.Equal(math.LegacyNewDec(-1)) {
		return fmt.Errorf("invalid validator bond factor: %s", v)
	}

	return nil
}

func validateGlobalLiquidStakingCap(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("global liquid staking cap cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("global liquid staking cap cannot be greater than 100%%: %s", v)
	}

	return nil
}

func validateValidatorLiquidStakingCap(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("validator liquid staking cap cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("validator liquid staking cap cannot be greater than 100%%: %s", v)
	}

	return nil
}
