package types

import "github.com/cosmos/cosmos-sdk/x/params/subspace"

const (
	// ModuleKey defines the name of the module
	ModuleName = "params"

	// RouterKey defines the routing key for a ParameterChangeProposal
	RouterKey = "params"

	// StoreKey is the string key for the params store
	StoreKey = subspace.StoreKey

	// TStoreKey is the string key for the params transient store
	TStoreKey = subspace.TStoreKey
)
