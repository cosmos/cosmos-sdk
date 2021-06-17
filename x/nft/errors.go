package nft

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrInvalidCollection     = sdkerrors.Register(ModuleName, 2, "invalid nft collection")
	ErrUnknownCollection     = sdkerrors.Register(ModuleName, 3, "unknown nft collection")
	ErrInvalidNFT            = sdkerrors.Register(ModuleName, 4, "invalid nft")
	ErrNFTAlreadyExists      = sdkerrors.Register(ModuleName, 5, "nft already exists")
	ErrUnknownNFT            = sdkerrors.Register(ModuleName, 6, "unknown nft")
	ErrEmptyTokenData        = sdkerrors.Register(ModuleName, 7, "nft data can't be empty")
	ErrUnauthorized          = sdkerrors.Register(ModuleName, 8, "unauthorized address")
	ErrInvalidTokenID        = sdkerrors.Register(ModuleName, 10, "invalid nft id")
	ErrInvalidTokenURI       = sdkerrors.Register(ModuleName, 11, "invalid nft uri")
	ErrInvalidCollectionName = sdkerrors.Register(ModuleName, 12, "invalid collection name")
)
