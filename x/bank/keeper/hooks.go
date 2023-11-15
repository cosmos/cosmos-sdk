package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Implements SendHooks interface
var _ types.SendHooks = BaseSendKeeper{}

func (keeper BaseSendKeeper) AfterSend(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if keeper.hooks != nil {
		err := keeper.hooks.AfterSend(ctx, fromAddr, toAddr, amt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (keeper BaseSendKeeper) BeforeSend(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if keeper.hooks != nil {
		err := keeper.hooks.BeforeSend(ctx, fromAddr, toAddr, amt)
		if err != nil {
			return err
		}
	}
	return nil

}

func (keeper BaseSendKeeper) BeforeMultiSend(ctx context.Context, input types.Input, outputs []types.Output) error {
	if keeper.hooks != nil {
		err := keeper.hooks.BeforeMultiSend(ctx, input, outputs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (keeper BaseSendKeeper) AfterMultiSend(ctx context.Context, input types.Input, outputs []types.Output) error {
	if keeper.hooks != nil {
		err := keeper.hooks.AfterMultiSend(ctx, input, outputs)
		if err != nil {
			return err
		}
	}
	return nil
}
