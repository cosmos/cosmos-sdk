package testutil

import (
	"context"

	"cosmossdk.io/core/address"
	banktypes "cosmossdk.io/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	AddressCodec() address.Codec
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper extends bank `MsgServer` to mock `Send` and to register handlers in MsgServiceRouter
type BankKeeper interface {
	banktypes.MsgServer

	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}
