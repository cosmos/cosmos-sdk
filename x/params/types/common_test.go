package types_test

import (

	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/types"
	prototypes "github.com/gogo/protobuf/types"
)

var (
	keyUnbondingTime = []byte("UnbondingTime")
	keyMaxValidators = []byte("MaxValidators")
	keyBondDenom     = []byte("BondDenom")

	key  = sdk.NewKVStoreKey("storekey")
	tkey = sdk.NewTransientStoreKey("transientstorekey")
)

type params struct {
	UnbondingTime prototypes.Duration `json:"unbonding_time" yaml:"unbonding_time"`
	MaxValidators prototypes.UInt64Value `json:"max_validators" yaml:"max_validators"`
	BondDenom     prototypes.StringValue `json:"bond_denom" yaml:"bond_denom"`
}

func validateUnbondingTime(i proto.Message) error {
	v, ok := i.(*prototypes.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	dur, err := prototypes.DurationFromProto(v)
	if err != nil {
		return err
	}
	if dur < (24 * time.Hour) {
		return fmt.Errorf("unbonding time must be at least one day")
	}

	return nil
}

func validateMaxValidators(i proto.Message) error {
	_, ok := i.(*prototypes.UInt64Value)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBondDenom(i proto.Message) error {
	v, ok := i.(*prototypes.StringValue)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}


	if len(v.Value) == 0 {
		return errors.New("denom cannot be empty")
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
