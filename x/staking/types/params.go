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
	// DefaultUnbondingTime reflects three weeks in seconds as the default
	// unbonding time.
	// TODO: Justify our choice of default here.
	DefaultUnbondingTime time.Duration = time.Second * 60 * 60 * 24 * 3

	// Default maximum number of bonded validators
	DefaultMaxValidators uint16 = 100

	// Default maximum entries in a UBD/RED pair
	DefaultMaxEntries uint16 = 7
)

// nolint - Keys for parameter access
var (
	KeyUnbondingTime = []byte("UnbondingTime")
	KeyMaxValidators = []byte("MaxValidators")
	KeyMaxEntries    = []byte("KeyMaxEntries")
	KeyBondDenom     = []byte("BondDenom")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for staking
type Params struct {
	UnbondingTime time.Duration `json:"unbonding_time"` // time duration of unbonding
	MaxValidators uint16        `json:"max_validators"` // maximum number of validators (max uint16 = 65535)
	MaxEntries    uint16        `json:"max_entries"`    // max entries for either unbonding delegation or redelegation (per pair/trio)
	// note: we need to be a bit careful about potential overflow here, since this is user-determined
	BondDenom string `json:"bond_denom"` // bondable coin denomination
}

func NewParams(unbondingTime time.Duration, maxValidators, maxEntries uint16,
	bondDenom string) Params {

	return Params{
		UnbondingTime: unbondingTime,
		MaxValidators: maxValidators,
		MaxEntries:    maxEntries,
		BondDenom:     bondDenom,
	}
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyUnbondingTime, &p.UnbondingTime},
		{KeyMaxValidators, &p.MaxValidators},
		{KeyMaxEntries, &p.MaxEntries},
		{KeyBondDenom, &p.BondDenom},
	}
}

// Equal returns a boolean determining if two Param types are identical.
// TODO: This is slower than comparing struct fields directly
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(DefaultUnbondingTime, DefaultMaxValidators, DefaultMaxEntries, sdk.DefaultBondDenom)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  Unbonding Time:    %s
  Max Validators:    %d
  Max Entries:       %d
  Bonded Coin Denom: %s`, p.UnbondingTime,
		p.MaxValidators, p.MaxEntries, p.BondDenom)
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
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &params)
	if err != nil {
		return
	}
	return
}

// validate a set of params
func (p Params) Validate() error {
	if p.BondDenom == "" {
		return fmt.Errorf("staking parameter BondDenom can't be an empty string")
	}
	if p.MaxValidators == 0 {
		return fmt.Errorf("staking parameter MaxValidators must be a positive integer")
	}
	return nil
}
