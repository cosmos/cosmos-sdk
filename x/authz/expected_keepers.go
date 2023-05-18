package authz

import (
	context "context"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	address.Codec

	GetAccount(ctx context.Context, addr sdk.AccAddress) (sdk.AccountI,error)
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) (sdk.AccountI,error)
	SetAccount(ctx context.Context, acc sdk.AccountI)error
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
}
