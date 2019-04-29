// nolint
package nfts

import (
	"github.com/cosmos/cosmos-sdk/x/nfts/keeper"
	"github.com/cosmos/cosmos-sdk/x/nfts/types"
)

type (
	Keeper       = keeper.Keeper
	NFT          = types.NFT
	Collection   = types.Collection
	GenesisState = types.GenesisState
)

var (
	NewKeeper          = keeper.NewKeeper
	RegisterInvariants = keeper.RegisterInvariants

	NewNFT              = types.NewNFT
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
)

const (
	StoreKey     = keeper.StoreKey
	QuerierRoute = keeper.QuerierRoute
)
