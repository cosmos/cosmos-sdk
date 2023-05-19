package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ SendCoinsHooks = &MultiSendCoinsHooks{}

type MultiSendCoinsHooks []SendCoinsHooks

func NewMultiSendCoinsHooks(hooks ...SendCoinsHooks) MultiSendCoinsHooks {
	return hooks
}

func (h MultiSendCoinsHooks) AfterSendCoins(
	ctx context.Context,
	fromAddr sdk.AccAddress,
	toAddr sdk.AccAddress,
	amount sdk.Coins,
) error {
	for i := range h {
		if err := h[i].AfterSendCoins(ctx, fromAddr, toAddr, amount); err != nil {
			return err
		}
	}

	return nil
}
