package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ SendCoinsHooks = &MultiSendCoinsHooks{}

type MultiSendCoinsHooks []SendCoinsHooks

func NewMultiSendCoinsHooks(hooks ...SendCoinsHooks) MultiSendCoinsHooks {
	return hooks
}

func (h MultiSendCoinsHooks) BeforeSendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount sdk.Coins) error {
	for i := range h {
		if err := h[i].BeforeSendCoins(ctx, fromAddr, toAddr, amount); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiSendCoinsHooks) AfterSendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount sdk.Coins) error {
	for i := range h {
		if err := h[i].AfterSendCoins(ctx, fromAddr, toAddr, amount); err != nil {
			return err
		}
	}

	return nil
}
