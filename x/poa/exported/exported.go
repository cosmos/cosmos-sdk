package exported

import (
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DelegationI delegation bond for a delegated proof of stake system
type DelegationI interface {
	GetValidatorAddr() sdk.ValAddress // validator operator address
	GetShares() sdk.Dec               // amount of validator's shares held in this delegation
}

// ValidatorI expected validator functions
type ValidatorI interface {
	IsJailed() bool               // whether the validator is jailed
	GetMoniker() string           // moniker of the validator
	GetStatus() sdk.BondStatus    // status of the validator
	IsBonded() bool               // check if has a bonded status
	IsUnbonded() bool             // check if has status unbonded
	IsUnbonding() bool            // check if has status unbonding
	GetOperator() sdk.ValAddress  // operator address to receive/return validators coins
	GetConsPubKey() crypto.PubKey // validation consensus pubkey
	GetConsAddr() sdk.ConsAddress // validation consensus address
	GetWeight() sdk.Int           // validation weight
	GetBondedWeight() sdk.Int     // validator bonded weight
	GetConsensusPower() int64     // validation power in tendermint
}
