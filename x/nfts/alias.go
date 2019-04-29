// nolint
package nfts

import (
	"github.com/cosmos/cosmos-sdk/x/nfts/keeper"
	"github.com/cosmos/cosmos-sdk/x/nfts/types"
)

type (
	Keeper             = keeper.Keeper
	NFT                = types.NFT
	NFTs               = types.NFTs
	Collection         = types.Collection
	Collections        = types.Collections
	GenesisState       = types.GenesisState
	MsgTransferNFT     = types.MsgTransferNFT
	MsgEditNFTMetadata = types.MsgEditNFTMetadata
	MsgMintNFT         = types.MsgMintNFT
	MsgBurnNFT         = types.MsgBurnNFT
	MsgBuyNFT          = types.MsgBuyNFT
)

var (
	NewKeeper          = keeper.NewKeeper
	RegisterInvariants = keeper.RegisterInvariants

	NewNFT              = types.NewNFT
	NewNFTs             = types.NewNFTs
	NewCollection       = types.NewCollection
	NewCollections      = types.NewCollections
	EmptyCollection     = types.EmptyCollection
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
)

const (
	StoreKey     = keeper.StoreKey
	QuerierRoute = keeper.QuerierRoute
)
