package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// ExternalCommunityPoolKeeper is the interface that an external community pool module keeper must fulfill
// for x/distribution to properly accept it as a community pool fund destination.
type ExternalCommunityPoolKeeper interface {
	// GetCommunityPoolModule gets the module name that funds should be sent to for the community pool.
	// This is the address that x/distribution will send funds to for external management.
	GetCommunityPoolModule() string
	// FundCommunityPool allows an account to directly fund the community fund pool.
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, senderAddr sdk.AccAddress) error
	// DistributeFromCommunityPool distributes funds from the community pool module account to
	// a receiver address.
	DistributeFromCommunityPool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error
}
