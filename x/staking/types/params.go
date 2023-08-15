package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
	// ValidatorBondFactor of -1 indicates that it's disabled
	ValidatorBondCapDisabled = sdk.NewDecFromInt(sdk.NewInt(-1))

	// DefaultValidatorBondFactor is set to -1 (disabled)
	DefaultValidatorBondFactor = ValidatorBondCapDisabled
	// DefaultGlobalLiquidStakingCap is set to 100%
	DefaultGlobalLiquidStakingCap = sdk.OneDec()
	// DefaultValidatorLiquidStakingCap is set to 100%
	DefaultValidatorLiquidStakingCap = sdk.OneDec()
)

var (
	KeyUnbondingTime             = []byte("UnbondingTime")
	KeyMaxValidators             = []byte("MaxValidators")
	KeyMaxEntries                = []byte("MaxEntries")
	KeyBondDenom                 = []byte("BondDenom")
	KeyHistoricalEntries         = []byte("HistoricalEntries")
	KeyPowerReduction            = []byte("PowerReduction")
	KeyValidatorBondFactor       = []byte("ValidatorBondFactor")
	KeyGlobalLiquidStakingCap    = []byte("GlobalLiquidStakingCap")
	KeyValidatorLiquidStakingCap = []byte("ValidatorLiquidStakingCap")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamTable for staking module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	unbondingTime time.Duration,
	maxValidators uint32,
	maxEntries uint32,
	historicalEntries uint32,
	bondDenom string,
	validatorBondFactor sdk.Dec,
	globalLiquidStakingCap sdk.Dec,
	validatorLiquidStakingCap sdk.Dec,
) Params {
	return Params{
		UnbondingTime:             unbondingTime,
		MaxValidators:             maxValidators,
		MaxEntries:                maxEntries,
		HistoricalEntries:         historicalEntries,
		BondDenom:                 bondDenom,
		ValidatorBondFactor:       validatorBondFactor,
		GlobalLiquidStakingCap:    globalLiquidStakingCap,
		ValidatorLiquidStakingCap: validatorLiquidStakingCap,
	}
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyUnbondingTime, &p.UnbondingTime, validateUnbondingTime),
		paramtypes.NewParamSetPair(KeyMaxValidators, &p.MaxValidators, validateMaxValidators),
		paramtypes.NewParamSetPair(KeyMaxEntries, &p.MaxEntries, validateMaxEntries),
		paramtypes.NewParamSetPair(KeyHistoricalEntries, &p.HistoricalEntries, validateHistoricalEntries),
		paramtypes.NewParamSetPair(KeyBondDenom, &p.BondDenom, validateBondDenom),
		paramtypes.NewParamSetPair(KeyValidatorBondFactor, &p.ValidatorBondFactor, validateValidatorBondFactor),
		paramtypes.NewParamSetPair(KeyGlobalLiquidStakingCap, &p.GlobalLiquidStakingCap, validateGlobalLiquidStakingCap),
		paramtypes.NewParamSetPair(KeyValidatorLiquidStakingCap, &p.ValidatorLiquidStakingCap, validateValidatorLiquidStakingCap),
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
		DefaultValidatorBondFactor,
		DefaultGlobalLiquidStakingCap,
		DefaultValidatorLiquidStakingCap,
	)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
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
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.NewInt(1)) {
		return fmt.Errorf("power reduction cannot be lower than 1")
	}

	return nil
}

func validateValidatorBondFactor(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() && !v.Equal(sdk.NewDec(-1)) {
		return fmt.Errorf("invalid validator bond factor: %s", v)
	}

	return nil
}

func validateGlobalLiquidStakingCap(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("global liquid staking cap cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("global liquid staking cap cannot be greater than 100%%: %s", v)
	}

	return nil
}

func validateValidatorLiquidStakingCap(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("validator liquid staking cap cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("validator liquid staking cap cannot be greater than 100%%: %s", v)
	}

	return nil
}
