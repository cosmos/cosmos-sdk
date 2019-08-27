package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
	yaml "gopkg.in/yaml.v2"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Validator struct {
	OperatorAddress         sdk.ValAddress           `json:"operator_address" yaml:"operator_address"` // address of the validator's operator; bech encoded in JSON
	ConsPubKey              crypto.PubKey            `json:"consensus_pubkey" yaml:"consensus_pubkey"` // the consensus public key of the validator; bech encoded in JSON
	Jailed                  bool                     `json:"jailed" yaml:"jailed"`                     // has the validator been jailed from bonded status?
	Status                  sdk.BondStatus           `json:"status" yaml:"status"`                     // validator status (bonded/unbonding/unbonded)
	Weight                  sdk.Int                  `json:"weight" yaml:"weight"`                     // weight (power) associated to a validator
	Description             stakingtypes.Description `json:"description" yaml:"description"`           // description terms for the validator
	UnbondingHeight         int64                    `json:"unbonding_height" yaml:"unbonding_height"` // if unbonding, height at which this validator has begun unbonding
	UnbondingCompletionTime time.Time                `json:"unbonding_time" yaml:"unbonding_time"`     // if unbonding, min time for the validator to complete unbonding
}

// NewValidator - create a new Validator
func NewValidator(operator sdk.ValAddress, pubKey crypto.PubKey, description stakingtypes.Description) Validator {
	return Validator{
		OperatorAddress:         operator,
		ConsPubKey:              pubKey,
		Jailed:                  false,
		Status:                  sdk.Unbonded,
		Weight:                  sdk.NewInt(int64(10)), // set default to 10 when creating a validator
		Description:             description,
		UnbondingHeight:         int64(0),
		UnbondingCompletionTime: time.Unix(0, 0).UTC(),
	}
}

// custom marshal yaml function due to consensus pubkey
func (v Validator) MarshalYAML() (interface{}, error) {
	bs, err := yaml.Marshal(bechValidator{
		OperatorAddress:         v.OperatorAddress,
		ConsPubKey:              sdk.MustBech32ifyConsPub(v.ConsPubKey),
		Jailed:                  v.Jailed,
		Status:                  v.Status,
		Weight:                  v.Weight,
		Description:             v.Description,
		UnbondingHeight:         v.UnbondingHeight,
		UnbondingCompletionTime: v.UnbondingCompletionTime,
	})
	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

// String returns a human readable string representation of a validator.
func (v Validator) String() string {
	bechConsPubKey, err := sdk.Bech32ifyConsPub(v.ConsPubKey)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`Validator
  Operator Address:           %s
  Validator Consensus Pubkey: %s
  Jailed:                     %v
  Status:                     %s
  Weight:                     %s
  Description:                %s
  Unbonding Height:           %d
  Unbonding Completion Time:  %v`, v.OperatorAddress, bechConsPubKey,
		v.Jailed, v.Status, v.Weight,
		v.Description, v.UnbondingHeight,
		v.UnbondingCompletionTime)
}

func MustMarshalValidator(cdc *codec.Codec, val Validator) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(val)
}

// unmarshal a redelegation from a store value
func MustUnmarshalValidator(cdc *codec.Codec, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}
	return validator
}

// unmarshal a redelegation from a store value
func UnmarshalValidator(cdc *codec.Codec, value []byte) (validator Validator, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &validator)
	return validator, err
}

// this is a helper struct used for JSON de- and encoding only
type bechValidator struct {
	OperatorAddress         sdk.ValAddress           `json:"operator_address" yaml:"operator_address"` // address of the validator's operator; bech encoded in JSON
	ConsPubKey              string                   `json:"consensus_pubkey" yaml:"consensus_pubkey"` // the consensus public key of the validator; bech encoded in JSON
	Jailed                  bool                     `json:"jailed" yaml:"jailed"`                     // has the validator been jailed from bonded status?
	Status                  sdk.BondStatus           `json:"status" yaml:"status"`                     // validator status (bonded/unbonding/unbonded)
	Weight                  sdk.Int                  `json:"weight" yaml:"weight"`                     // weight (power) associated to a validator
	Description             stakingtypes.Description `json:"description" yaml:"description"`           // description terms for the validator
	UnbondingHeight         int64                    `json:"unbonding_height" yaml:"unbonding_height"` // if unbonding, height at which this validator has begun unbonding
	UnbondingCompletionTime time.Time                `json:"unbonding_time" yaml:"unbonding_time"`     // if unbonding, min time for the validator to complete unbonding
}

// MarshalJSON marshals the validator to JSON using Bech32
func (v Validator) MarshalJSON() ([]byte, error) {
	bechConsPubKey, err := sdk.Bech32ifyConsPub(v.ConsPubKey)
	if err != nil {
		return nil, err
	}

	return codec.Cdc.MarshalJSON(bechValidator{
		OperatorAddress:         v.OperatorAddress,
		ConsPubKey:              bechConsPubKey,
		Jailed:                  v.Jailed,
		Status:                  v.Status,
		Weight:                  v.Weight,
		Description:             v.Description,
		UnbondingHeight:         v.UnbondingHeight,
		UnbondingCompletionTime: v.UnbondingCompletionTime,
	})
}

// UnmarshalJSON unmarshals the validator from JSON using Bech32
func (v *Validator) UnmarshalJSON(data []byte) error {
	bv := &bechValidator{}
	if err := codec.Cdc.UnmarshalJSON(data, bv); err != nil {
		return err
	}
	consPubKey, err := sdk.GetConsPubKeyBech32(bv.ConsPubKey)
	if err != nil {
		return err
	}
	*v = Validator{
		OperatorAddress:         bv.OperatorAddress,
		ConsPubKey:              consPubKey,
		Jailed:                  bv.Jailed,
		Weight:                  bv.Weight,
		Status:                  bv.Status,
		Description:             bv.Description,
		UnbondingHeight:         bv.UnbondingHeight,
		UnbondingCompletionTime: bv.UnbondingCompletionTime,
	}
	return nil
}

// only the vitals
func (v Validator) TestEquivalent(v2 Validator) bool {
	return v.ConsPubKey.Equals(v2.ConsPubKey) &&
		bytes.Equal(v.OperatorAddress, v2.OperatorAddress) &&
		v.Status.Equal(v2.Status) &&
		v.Weight.Equal(v2.Weight) &&
		v.Description == v2.Description
}

// return the TM validator address
func (v Validator) ConsAddress() sdk.ConsAddress {
	return sdk.ConsAddress(v.ConsPubKey.Address())
}

// IsBonded checks if the validator status equals Bonded
func (v Validator) IsBonded() bool {
	return v.GetStatus().Equal(sdk.Bonded)
}

// IsUnbonded checks if the validator status equals Unbonded
func (v Validator) IsUnbonded() bool {
	return v.GetStatus().Equal(sdk.Unbonded)
}

// IsUnbonding checks if the validator status equals Unbonding
func (v Validator) IsUnbonding() bool {
	return v.GetStatus().Equal(sdk.Unbonding)
}

// ABCIValidatorUpdate returns an abci.ValidatorUpdate from a staking validator type
// with the full validator power
func (v Validator) ABCIValidatorUpdate() abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		PubKey: tmtypes.TM2PB.PubKey(v.ConsPubKey),
		Power:  v.ConsensusPower(),
	}
}

// ABCIValidatorUpdateZero returns an abci.ValidatorUpdate from a staking validator type
// with zero power used for validator updates.
func (v Validator) ABCIValidatorUpdateZero() abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		PubKey: tmtypes.TM2PB.PubKey(v.ConsPubKey),
		Power:  0,
	}
}

// get the consensus-engine power
// a reduction of 10^6 from validator weight is applied
func (v Validator) ConsensusPower() int64 {
	if v.IsBonded() {
		return v.PotentialConsensusPower()
	}
	return 0
}

// potential consensus-engine power
func (v Validator) PotentialConsensusPower() int64 {
	return sdk.TokensToConsensusPower(v.Weight) // this is based off of weight
}

// UpdateStatus updates the location of the shares within a validator
// to reflect the new status
func (v Validator) UpdateStatus(newStatus sdk.BondStatus) Validator {
	v.Status = newStatus
	return v
}

// RemoveTokens removes weight (power) from a validator
func (v Validator) RemoveWeight(weight sdk.Int) Validator {
	if weight.IsNegative() {
		panic(fmt.Sprintf("should not happen: trying to remove negative weight %v", weight))
	}
	if v.Weight.LT(weight) {
		panic(fmt.Sprintf("should not happen: only have %v weight, trying to remove %v", v.Weight, weight))
	}
	v.Weight = v.Weight.Sub(weight)
	return v
}

// AddTokensFromDel adds tokens to a validator
func (v Validator) AddWeight(amount sdk.Int) Validator {
	v.Weight = v.Weight.Add(amount)
	return v
}

func (v Validator) GetBondedWeight() sdk.Int {
	if v.IsBonded() {
		return v.Weight
	}
	return sdk.ZeroInt()
}

// Validators is a collection of Validator
type Validators []Validator

func (v Validators) String() (out string) {
	for _, val := range v {
		out += val.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// func (v Validators) ToSDKValidators()

func (v Validator) IsJailed() bool               { return v.Jailed }
func (v Validator) GetMoniker() string           { return v.Description.Moniker }
func (v Validator) GetStatus() sdk.BondStatus    { return v.Status }
func (v Validator) GetOperator() sdk.ValAddress  { return v.OperatorAddress }
func (v Validator) GetConsPubKey() crypto.PubKey { return v.ConsPubKey }
func (v Validator) GetConsAddr() sdk.ConsAddress { return sdk.ConsAddress(v.ConsPubKey.Address()) }
func (v Validator) GetWeight() sdk.Int           { return v.Weight }
func (v Validator) GetConsensusPower() int64     { return v.ConsensusPower() }
