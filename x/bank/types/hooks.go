package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SendHooks interface {
	BeforeSend(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount sdk.Coins) error
	AfterSend(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount sdk.Coins) error
}

type MultiSendHooks []SendHooks

func NewMultiSendHooks(hooks ...SendHooks) MultiSendHooks {
	return hooks
}

func (h MultiSendHooks) BeforeSend(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount sdk.Coins) error {
	for i := range h {
		if err := h[i].BeforeSend(ctx, fromAddr, toAddr, amount); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiSendHooks) AfterSend(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount sdk.Coins) error {
	for i := range h {
		if err := h[i].AfterSend(ctx, fromAddr, toAddr, amount); err != nil {
			return err
		}
	}

	return nil
}
