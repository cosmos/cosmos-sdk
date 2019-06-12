// nolint
package nft

import (
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

type (
	Keeper             = keeper.Keeper
	NFT                = types.NFT
	BaseNFT            = types.BaseNFT
	NFTs               = types.NFTs
	Owner              = types.Owner
	IDCollection       = types.IDCollection
	Collection         = types.Collection
	Collections        = types.Collections
	GenesisState       = types.GenesisState
	MsgTransferNFT     = types.MsgTransferNFT
	MsgEditNFTMetadata = types.MsgEditNFTMetadata
)

var (
	NewKeeper             = keeper.NewKeeper
	RegisterInvariants    = keeper.RegisterInvariants
	NewQuerier            = keeper.NewQuerier
	NewBaseNFT            = types.NewBaseNFT
	NewNFTs               = types.NewNFTs
	NewIDCollection       = types.NewIDCollection
	NewOwner              = types.NewOwner
	NewCollection         = types.NewCollection
	NewCollections        = types.NewCollections
	EmptyCollection       = types.EmptyCollection
	NewGenesisState       = types.NewGenesisState
	NewMsgTransferNFT     = types.NewMsgTransferNFT
	NewMsgEditNFTMetadata = types.NewMsgEditNFTMetadata
	DefaultGenesisState   = types.DefaultGenesisState
	ValidateGenesis       = types.ValidateGenesis
	RegisterCodec         = types.RegisterCodec
	ModuleCdc             = types.ModuleCdc
)

const (
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
	ModuleName   = types.ModuleName
)
