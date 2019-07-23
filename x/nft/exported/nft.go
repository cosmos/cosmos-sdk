package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NFT non fungible token interface
type NFT interface {
	GetID() string
	GetOwner() sdk.AccAddress
	SetOwner(address sdk.AccAddress)
	GetName() string
	GetDescription() string
	GetImage() string
	GetTokenURI() string
	EditMetadata(name, description, image, tokenURI string)
	String() string
}
