package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SupplyKeeper defines the expected supply manager (noalias)
type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
