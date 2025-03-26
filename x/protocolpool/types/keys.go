package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the module name constant used in many places
	//
	// The module account for this name is the x/protocolpool community pool.
	// It can recieve funds from distribution from users who use FundCommunityPool or from the ProtocolPoolDistrAccount.
	ModuleName = "protocolpool"

	// ProtocolPoolDistrAccount is an intermediary account that holds the funds to be distributed to the protocolpool accounts.
	//
	// It can recieve funds from `x/distribution` and distribute them to continuous funds and the community pool.
	ProtocolPoolDistrAccount = "protocolpool_distr"

	// StoreKey is the store key string for protocolpool
	StoreKey = ModuleName

	// RouterKey is the message route for protocolpool
	RouterKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a cyclic dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/b62a28aac041829da5ded4aeacfcd7a42873d1c8/x/gov/types/keys.go#L9
	GovModuleName = "gov" // TODO this should be distribution
)

var (
	ContinuousFundsKey = collections.NewPrefix(3)
	ParamsKey          = collections.NewPrefix(8)
)
