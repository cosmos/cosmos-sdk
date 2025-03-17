package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple send hooks, all hook functions are run in array sequence
var _ SendHooks = &MultiSendHooks{}

type MultiSendHooks []SendHooks

func NewMultiSendHooks(hooks ...SendHooks) MultiSendHooks {
	return hooks
}

func (h MultiSendHooks) BeforeSend(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	for i := range h {
		if err := h[i].BeforeSend(ctx, fromAddr, toAddr, amt); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiSendHooks) AfterSend(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	for i := range h {
		if err := h[i].AfterSend(ctx, fromAddr, toAddr, amt); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiSendHooks) BeforeMultiSend(ctx context.Context, inputs []Input, outputs []Output) error {
	for i := range h {
		if err := h[i].BeforeMultiSend(ctx, inputs, outputs); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiSendHooks) AfterMultiSend(ctx context.Context, inputs []Input, outputs []Output) error {
	for i := range h {
		if err := h[i].AfterMultiSend(ctx, inputs, outputs); err != nil {
			return err
		}
	}

	return nil
}
