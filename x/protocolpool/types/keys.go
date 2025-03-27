package types

import (
	"cosmossdk.io/collections"
)

const (
	// ModuleName is the module name constant used in many places
	//
	// The module account associated with this name is the x/protocolpool community pool module account.
	// It can receive funds from distribution from users who use FundCommunityPool or from the ProtocolPoolDistrAccount.
	ModuleName = "protocolpool"

	// ProtocolPoolDistrAccount is an intermediary account that holds the funds to be distributed to the protocolpool accounts.
	//
	// It can receive funds from `x/distribution` and distribute them to continuous funds and the community pool.
	ProtocolPoolDistrAccount = "protocolpool_distr"

	// StoreKey is the store key string for protocolpool
	StoreKey = ModuleName

	// RouterKey is the message route for protocolpool
	RouterKey = ModuleName
)

var (
	ContinuousFundsKey = collections.NewPrefix(3)
	ParamsKey          = collections.NewPrefix(8)
)
