package nft

import (
	"cosmossdk.io/errors"
)

// x/nft module sentinel errors
var (
	ErrClassExists    = errors.Register(ModuleName, 2, "nft class already exist")
	ErrClassNotExists = errors.Register(ModuleName, 3, "nft class does not exist")
	ErrNFTExists      = errors.Register(ModuleName, 4, "nft already exist")
	ErrNFTNotExists   = errors.Register(ModuleName, 5, "nft does not exist")
	ErrEmptyClassID   = errors.Register(ModuleName, 6, "empty class id")
	ErrEmptyNFTID     = errors.Register(ModuleName, 7, "empty nft id")
)
