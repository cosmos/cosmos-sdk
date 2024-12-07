package types

import (
	"context"

	"cosmossdk.io/core/transaction"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountsModKeeper defines the contract for x/accounts APIs
type AccountsModKeeper interface {
	// Query is used to query an account
	Query(
		ctx context.Context,
		accountAddr []byte,
		queryRequest transaction.Msg,
	) (transaction.Msg, error)

	Init(
		ctx context.Context,
		accountType string,
		creator []byte,
		initRequest transaction.Msg,
		funds sdk.Coins,
	) (transaction.Msg, []byte, error)

	Execute(
		ctx context.Context,
		accountAddr []byte,
		sender []byte,
		execRequest transaction.Msg,
		funds sdk.Coins,
	) (transaction.Msg, error)
}
