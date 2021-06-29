package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/nft module sentinel errors
var (
	ErrInvalidMetadata   = sdkerrors.Register(ModuleName, 2, "invalid metadata")
	ErrInvalidNFT        = sdkerrors.Register(ModuleName, 3, "invalid nft")
	ErrTypeExists        = sdkerrors.Register(ModuleName, 4, "nft type already exist")
	ErrTypeNotExists     = sdkerrors.Register(ModuleName, 5, "nft type not exist")
	ErrNFTExists         = sdkerrors.Register(ModuleName, 6, "nft already exist")
	ErrNFTNotExists      = sdkerrors.Register(ModuleName, 7, "nft not exist")
	ErrNFTEditRestricted = sdkerrors.Register(ModuleName, 8, "nft is defined as editing restricted")
)
