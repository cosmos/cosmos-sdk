package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/nft module sentinel errors
//
var (
	ErrNoNFTFound = sdkerrors.Register(ModuleName, 1, "nft does not exist")
)
