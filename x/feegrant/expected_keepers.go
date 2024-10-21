package feegrant

import (
	"context"

	"cosmossdk.io/core/transaction"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected supply Keeper (noalias)
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type AccountsKeeper interface {
	Execute(
		ctx context.Context,
		accountAddr []byte,
		sender []byte,
		execRequest transaction.Msg,
		funds sdk.Coins,
	) (transaction.Msg, error)
}
