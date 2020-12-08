package exported

import (
	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisBalance defines a genesis balance interface that allows for account
// address and balance retrieval.
type GenesisBalance interface {
	GetAddress() sdk.AccAddress
	GetCoins() sdk.Coins
}

// SupplyI defines an inflationary supply interface for modules that handle
// token supply.
type SupplyI interface {
	proto.Message

	GetTotal() sdk.Coins
	SetTotal(total sdk.Coins)

	Inflate(amount sdk.Coins)
	Deflate(amount sdk.Coins)

	String() string
	ValidateBasic() error
}
