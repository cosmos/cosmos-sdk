package types_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	keyUnbondingTime = []byte("UnbondingTime")
	keyMaxValidators = []byte("MaxValidators")
	keyBondDenom     = []byte("BondDenom")

	key  = sdk.NewKVStoreKey("storekey")
	tkey = sdk.NewTransientStoreKey("transientstorekey")
)

type params struct {
	UnbondingTime time.Duration `json:"unbonding_time" yaml:"unbonding_time"`
	MaxValidators uint16        `json:"max_validators" yaml:"max_validators"`
	BondDenom     string        `json:"bond_denom" yaml:"bond_denom"`
}

func validateUnbondingTime(currentValue, newValue interface{}) error {
	newUbdTime, ok := newValue.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", newValue)
	}

	if newUbdTime < (24 * time.Hour) {
		return fmt.Errorf("unbonding time must be at least one day")
	}

	// ignore error if there is no previous value
	if currentValue == nil {
		return nil
	}

	currentUbdTime, ok := currentValue.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", currentValue)
	}

	if currentUbdTime == newUbdTime {
		return fmt.Errorf("new unbonding time cannot be the same as the current value %s, %s", currentUbdTime, newUbdTime)
	}

	return nil
}

func validateMaxValidators(currentValue, newValue interface{}) error {
	newVals, ok := newValue.(uint16)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", newValue)
	}

	// ignore error if there is no previous value
	if currentValue == nil {
		return nil
	}

	currentVals, ok := currentValue.(uint16)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", currentValue)
	}

	if currentVals == newVals {
		return fmt.Errorf("proposed max validators has the same value as the current one, %d = %d", currentVals, newVals)
	}

	return nil
}

func validateBondDenom(currentValue, newValue interface{}) error {
	newDenom, ok := newValue.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", newValue)
	}

	if err := sdk.ValidateDenom(newDenom); err != nil {
		return err
	}

	// ignore error if there is no previous value
	if currentValue == nil {
		return nil
	}

	currentDenom, ok := currentValue.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", currentValue)
	}

	if currentDenom == newDenom {
		return fmt.Errorf("proposed bond denomination cannot be the same as the current one, %s = %s", newDenom, currentDenom)
	}

	return nil
}

func (p *params) ParamSetPairs() types.ParamSetPairs {
	return types.ParamSetPairs{
		{keyUnbondingTime, &p.UnbondingTime, validateUnbondingTime},
		{keyMaxValidators, &p.MaxValidators, validateMaxValidators},
		{keyBondDenom, &p.BondDenom, validateBondDenom},
	}
}

func paramKeyTable() types.KeyTable {
	return types.NewKeyTable().RegisterParamSet(&params{})
}
