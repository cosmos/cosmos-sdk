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
		panic("improper use of BondStatusToString")
	}
}

// nolint
func (b BondStatus) Equal(b2 BondStatus) bool {
	return byte(b) == byte(b2)
}

// validator for a delegated proof of stake system
type Validator interface {
	GetJailed() bool          // whether the validator is jailed
	GetMoniker() string       // moniker of the validator
	GetStatus() BondStatus    // status of the validator
	GetOperator() ValAddress  // operator address to receive/return validators coins
	GetPubKey() crypto.PubKey // validation pubkey
	GetPower() Dec            // validation power
	GetTokens() Dec           // validation tokens
	GetDelegatorShares() Dec  // Total out standing delegator shares
	GetBondHeight() int64     // height in which the validator became active
}

// validator which fulfills abci validator interface for use in Tendermint
func ABCIValidator(v Validator) abci.Validator {
	return abci.Validator{
		PubKey:  tmtypes.TM2PB.PubKey(v.GetPubKey()),
		Address: v.GetPubKey().Address(),
		Power:   v.GetPower().RoundInt64(),
	}
}

// properties for the set of all validators
type ValidatorSet interface {
	// iterate through validators by operator address, execute func for each validator
	IterateValidators(Context,
		func(index int64, validator Validator) (stop bool))

	// iterate through bonded validators by operator address, execute func for each validator
	IterateValidatorsBonded(Context,
		func(index int64, validator Validator) (stop bool))

	Validator(Context, ValAddress) Validator            // get a particular validator by operator
	ValidatorByPubKey(Context, crypto.PubKey) Validator // get a particular validator by signing PubKey
	TotalPower(Context) Dec                             // total power of the validator set

	// slash the validator and delegators of the validator, specifying offence height, offence power, and slash fraction
	Slash(Context, crypto.PubKey, int64, int64, Dec)
	Jail(Context, crypto.PubKey)   // jail a validator
	Unjail(Context, crypto.PubKey) // unjail a validator

	// Delegation allows for getting a particular delegation for a given validator
	// and delegator outside the scope of the staking module.
	Delegation(Context, AccAddress, ValAddress) Delegation
}

//_______________________________________________________________________________

// delegation bond for a delegated proof of stake system
type Delegation interface {
	GetDelegator() AccAddress // delegator AccAddress for the bond
	GetValidator() ValAddress // validator operator address
	GetBondShares() Dec       // amount of validator's shares
}

// properties for the set of all delegations for a particular
type DelegationSet interface {
	GetValidatorSet() ValidatorSet // validator set for which delegation set is based upon

	// iterate through all delegations from one delegator by validator-AccAddress,
	//   execute func for each validator
	IterateDelegations(ctx Context, delegator AccAddress,
		fn func(index int64, delegation Delegation) (stop bool))
}

// validator event hooks
// These can be utilized to communicate between a staking keeper
// and another keeper which must take particular actions when
// validators are bonded and unbonded. The second keeper must implement
// this interface, which then the staking keeper can call.
type ValidatorHooks interface {
	OnValidatorBonded(ctx Context, address ConsAddress)         // Must be called when a validator is bonded
	OnValidatorBeginUnbonding(ctx Context, address ConsAddress) // Must be called when a validator begins unbonding
}
