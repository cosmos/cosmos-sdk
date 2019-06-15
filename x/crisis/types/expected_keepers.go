package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DistributionKeeper defines the expected distribution keeper (noalias)
type DistributionKeeper interface {
	DistributeFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) sdk.Error
}

// SupplyKeeper defines the expected supply keeper (noalias)
type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error
}
