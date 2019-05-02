package nft

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
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

	HandleMsgEditNFTMetadata = handler.HandleMsgEditNFTMetadata
	HandleMsgMintNFT         = handler.HandleMsgMintNFT
	HandleMsgBurnNFT         = handler.HandleMsgBurnNFT

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
