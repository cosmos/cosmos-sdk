package types

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// AccountsModKeeper defines the contract for x/accounts APIs
type AccountsModKeeper interface {
	SendModuleMessageUntyped(ctx context.Context, sender []byte, msg protoiface.MessageV1) (protoiface.MessageV1, error)
	IsAccountsModuleAccount(ctx context.Context, accountAddr []byte) bool
	NextAccountNumber(ctx context.Context) (accNum uint64, err error)

	// InitAccountNumberSeqUnsafe is use to set accounts module account number with value
	// of auth module current account number
	InitAccountNumberSeqUnsafe(ctx context.Context, currentAccNum uint64) error
}
