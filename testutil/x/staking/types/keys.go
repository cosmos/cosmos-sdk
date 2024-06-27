package types

import (
	"cosmossdk.io/collections"
)

const (
	// ModuleName is the name of the staking module
	ModuleName = "mstaking"

	StoreKey = "staking"

	NotBondedPoolName = "not_bonded_tokens_pool"
	BondedPoolName    = "bonded_tokens_pool"
)

var (
	// Keys for store prefixes
	ValidatorsKey           = collections.NewPrefix(33) // prefix for each key to a validator
	ValidatorsByConsAddrKey = collections.NewPrefix(34) // prefix for each key to a validator index, by pubkey
	DelegationKey           = collections.NewPrefix(49) // key for a delegation
	ParamsKey               = collections.NewPrefix(81) // prefix for parameters for module x/staking
)
