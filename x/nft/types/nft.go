package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NFTI is an interface used to store NFTs at a given id and owner.
type NFTI interface {
	GetId() string // can not return empty string.
	GetOwner() sdk.AccAddress
}
