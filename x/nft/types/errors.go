package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/nft module sentinel errors
var (
	ErrInvalidMetadata = sdkerrors.Register(ModuleName, 2, "invalid metadata")
	ErrInvalidNFT      = sdkerrors.Register(ModuleName, 3, "invalid nft")
	ErrNFTTypeExists   = sdkerrors.Register(ModuleName, 4, "nft type already exist")
	ErrNFTExists       = sdkerrors.Register(ModuleName, 5, "nft already exist")
)
