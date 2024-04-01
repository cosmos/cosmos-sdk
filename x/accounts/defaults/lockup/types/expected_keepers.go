package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StakingKeeper interface {
	// unbond token without waiting for unbonding period
	Unbond(context.Context, sdk.AccAddress, sdk.ValAddress, math.LegacyDec) (math.Int, error)
}

type BankKeeper interface {
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}
