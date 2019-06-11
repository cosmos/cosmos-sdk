package nft

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

type (
	Keeper             = keeper.Keeper
	NFT                = types.NFT
	BaseNFT            = types.BaseNFT
	NFTs               = types.NFTs
	Collection         = types.Collection
	Collections        = types.Collections
	GenesisState       = types.GenesisState
	MsgTransferNFT     = types.MsgTransferNFT
	MsgEditNFTMetadata = types.MsgEditNFTMetadata
)

var (
	NewKeeper           = keeper.NewKeeper
	RegisterInvariants  = keeper.RegisterInvariants
	NewQuerier          = keeper.NewQuerier
	NewBaseNFT          = types.NewBaseNFT
	NewNFTs             = types.NewNFTs
	NewCollection       = types.NewCollection
	NewCollections      = types.NewCollections
	EmptyCollection     = types.EmptyCollection
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis
	RegisterCodec       = types.RegisterCodec
	ModuleCdc           = types.ModuleCdc
)

const (
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
	ModuleName   = types.ModuleName
)
