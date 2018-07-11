package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

// status of a validator
type BondStatus byte

// nolint
const (
	Unbonded  BondStatus = 0x00
	Unbonding BondStatus = 0x01
	Bonded    BondStatus = 0x02
)

//BondStatusToString for pretty prints of Bond Status
func BondStatusToString(b BondStatus) string {
	switch b {
	case 0x00:
		return "Unbonded"
	case 0x01:
		return "Unbonding"
	case 0x02:
		return "Bonded"
	default:
		return ""
	}
}

// validator for a delegated proof of stake system
type Validator interface {
	GetRevoked() bool         // whether the validator is revoked
	GetMoniker() string       // moniker of the validator
	GetStatus() BondStatus    // status of the validator
	GetOwner() AccAddress     // owner AccAddress to receive/return validators coins
	GetPubKey() crypto.PubKey // validation pubkey
	GetPower() Rat            // validation power
	GetDelegatorShares() Rat  // Total out standing delegator shares
	GetBondHeight() int64     // height in which the validator became active
}

// validator which fulfills abci validator interface for use in Tendermint
func ABCIValidator(v Validator) abci.Validator {
	return abci.Validator{
		PubKey: tmtypes.TM2PB.PubKey(v.GetPubKey()),
		Power:  v.GetPower().RoundInt64(),
	}
}

// properties for the set of all validators
type ValidatorSet interface {
	// iterate through validator by owner-AccAddress, execute func for each validator
	IterateValidators(Context,
		func(index int64, validator Validator) (stop bool))

	// iterate through bonded validator by pubkey-AccAddress, execute func for each validator
	IterateValidatorsBonded(Context,
		func(index int64, validator Validator) (stop bool))

	Validator(Context, AccAddress) Validator // get a particular validator by owner AccAddress
	TotalPower(Context) Rat                  // total power of the validator set

	// slash the validator and delegators of the validator, specifying offence height, offence power, and slash fraction
	Slash(Context, crypto.PubKey, int64, int64, Rat)
	Revoke(Context, crypto.PubKey)   // revoke a validator
	Unrevoke(Context, crypto.PubKey) // unrevoke a validator
}

//_______________________________________________________________________________

// delegation bond for a delegated proof of stake system
type Delegation interface {
	GetDelegator() AccAddress // delegator AccAddress for the bond
	GetValidator() AccAddress // validator owner AccAddress for the bond
	GetBondShares() Rat       // amount of validator's shares
}

// properties for the set of all delegations for a particular
type DelegationSet interface {
	GetValidatorSet() ValidatorSet // validator set for which delegation set is based upon

	// iterate through all delegations from one delegator by validator-AccAddress,
	//   execute func for each validator
	IterateDelegations(ctx Context, delegator AccAddress,
		fn func(index int64, delegation Delegation) (stop bool))
}
