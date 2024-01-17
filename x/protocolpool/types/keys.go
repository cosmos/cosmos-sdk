package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "protocolpool"

	// StreamAccount is the name constant used for stream account
	StreamAccount = "stream_acc"

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
	BudgetKey                    = collections.NewPrefix(2)
	ContinuousFundKey            = collections.NewPrefix(3)
	RecipientFundPercentageKey   = collections.NewPrefix(4)
	RecipientFundDistributionKey = collections.NewPrefix(5)
	ToDistributeKey              = collections.NewPrefix(6)
)
