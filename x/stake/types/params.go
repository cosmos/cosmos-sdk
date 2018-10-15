package types

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	// defaultUnbondingTime reflects three weeks in seconds as the default
	// unbonding time.
	defaultUnbondingTime time.Duration = 60 * 60 * 24 * 3 * time.Second

	// Delay, in blocks, between when validator updates are returned to Tendermint and when they are applied
	// For example, if this is 0, the validator set at the end of a block will sign the next block, or
	// if this is 1, the validator set at the end of a block will sign the block after the next.
	// Constant as this should not change without a hard fork.
	ValidatorUpdateDelay int64 = 1
)

// nolint - Keys for parameter access
var (
	KeyInflationRateChange = []byte("InflationRateChange")
	KeyInflationMax        = []byte("InflationMax")
	KeyInflationMin        = []byte("InflationMin")
	KeyGoalBonded          = []byte("GoalBonded")
	KeyUnbondingTime       = []byte("UnbondingTime")
	KeyMaxValidators       = []byte("MaxValidators")
	KeyBondDenom           = []byte("BondDenom")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for staking
type Params struct {
	InflationRateChange sdk.Dec `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Dec `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Dec `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Dec `json:"goal_bonded"`           // Goal of percent bonded atoms

	UnbondingTime time.Duration `json:"unbonding_time"`

	MaxValidators uint16 `json:"max_validators"` // maximum number of validators
	BondDenom     string `json:"bond_denom"`     // bondable coin denomination
}

// Implements params.ParamSet
func (p *Params) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{KeyInflationRateChange, &p.InflationRateChange},
		{KeyInflationMax, &p.InflationMax},
		{KeyInflationMin, &p.InflationMin},
		{KeyGoalBonded, &p.GoalBonded},
		{KeyUnbondingTime, &p.UnbondingTime},
		{KeyMaxValidators, &p.MaxValidators},
		{KeyBondDenom, &p.BondDenom},
	}
}

// Equal returns a boolean determining if two Param types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := MsgCdc.MustMarshalBinary(&p)
	bz2 := MsgCdc.MustMarshalBinary(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		InflationRateChange: sdk.NewDecWithPrec(13, 2),
		InflationMax:        sdk.NewDecWithPrec(20, 2),
		InflationMin:        sdk.NewDecWithPrec(7, 2),
		GoalBonded:          sdk.NewDecWithPrec(67, 2),
		UnbondingTime:       defaultUnbondingTime,
		MaxValidators:       100,
		BondDenom:           "steak",
	}
}

// HumanReadableString returns a human readable string representation of the
// parameters.
func (p Params) HumanReadableString() string {

	resp := "Pool \n"
	resp += fmt.Sprintf("Maximum Annual Inflation Rate Change: %s\n", p.InflationRateChange)
	resp += fmt.Sprintf("Max Inflation Rate: %s\n", p.InflationMax)
	resp += fmt.Sprintf("Min Inflation Tate: %s\n", p.InflationMin)
	resp += fmt.Sprintf("Bonded Token Goal (%s): %s\n", "s", p.GoalBonded)
	resp += fmt.Sprintf("Unbonding Time: %s\n", p.UnbondingTime)
	resp += fmt.Sprintf("Max Validators: %d: \n", p.MaxValidators)
	resp += fmt.Sprintf("Bonded Coin Denomination: %s\n", p.BondDenom)
	return resp
}

// unmarshal the current staking params value from store key or panic
func MustUnmarshalParams(cdc *codec.Codec, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}
	return params
}

// unmarshal the current staking params value from store key
func UnmarshalParams(cdc *codec.Codec, value []byte) (params Params, err error) {
	err = cdc.UnmarshalBinary(value, &params)
	if err != nil {
		return
	}
	return
}
