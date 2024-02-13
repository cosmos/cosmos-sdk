package types

import (
	"context"

	"cosmossdk.io/x/accounts/internal/implementation"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type AccountsModKeeper interface {
	SendAnyMessages(ctx context.Context, sender []byte, anyMessages []*implementation.Any) ([]*implementation.Any, error)
}
