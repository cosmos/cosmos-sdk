package nft

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/nft module sentinel errors
var (
	ErrInvalidNFT     = sdkerrors.Register(ModuleName, 2, "invalid nft")
	ErrClassExists    = sdkerrors.Register(ModuleName, 3, "nft class already exist")
	ErrClassNotExists = sdkerrors.Register(ModuleName, 4, "nft class does not exist")
	ErrNFTExists      = sdkerrors.Register(ModuleName, 5, "nft already exist")
	ErrNFTNotExists   = sdkerrors.Register(ModuleName, 6, "nft does not exist")
	ErrInvalidID      = sdkerrors.Register(ModuleName, 7, "invalid id")
	ErrInvalidClassID = sdkerrors.Register(ModuleName, 8, "invalid class id")
)
