package types

import (
	"cosmossdk.io/collections"
)

const (
	// ModuleName is the module name constant used in many places
	//
	// The module account associated with this name is the x/protocolpool community pool module account.
	// Funded by:
	// - direct funding from users using MsgFundProtocolPool
	// - all leftover funds that are not allocated to continuous funds in the BeginBlocker from ProtocolPoolEscrowAccount
	// Distributes to:
	// - users on MsgCommunityPoolSpend
	ModuleName = "protocolpool"

	// ProtocolPoolEscrowAccount is an intermediary account that holds the funds to be distributed to other accounts.
	//
	// It can receive funds from `x/distribution` and distribute them to continuous funds and/or the community pool.
	// Funded by:
	// - `x/distribution` during its BeginBlocker from the FeePool.CommunityPool
	// Distributes to:
	// - Continuous Funds in the store in the BeginBlocker
	// - CommunityPool (all remaining funds not allocated for Continuous Funds) in the BeginBlocker
	ProtocolPoolEscrowAccount = "protocolpool_escrow"

	// StoreKey is the store key string for protocolpool
	StoreKey = ModuleName
)

var (
	ContinuousFundsKey = collections.NewPrefix(3)
	ParamsKey          = collections.NewPrefix(8)
)
