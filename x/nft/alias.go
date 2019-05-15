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
	Balance            = types.Balance
	GenesisState       = types.GenesisState
	MsgTransferNFT     = types.MsgTransferNFT
	MsgEditNFTMetadata = types.MsgEditNFTMetadata
)

var (
	NewKeeper          = keeper.NewKeeper
	RegisterInvariants = keeper.RegisterInvariants

	NewBaseNFT          = types.NewBaseNFT
	NewNFTs             = types.NewNFTs
	NewCollection       = types.NewCollection
	NewCollections      = types.NewCollections
	EmptyCollection     = types.EmptyCollection
	NewBalance          = types.NewBalance
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis
	RegisterCodec       = types.RegisterCodec
	ModuleCdc           = types.ModuleCdc
)

const (
	StoreKey     = keeper.StoreKey
	QuerierRoute = keeper.QuerierRoute
	ModuleName   = keeper.ModuleName
)
