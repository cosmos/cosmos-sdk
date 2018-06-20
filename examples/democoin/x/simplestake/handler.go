package simplestake

import (
	abci "github.com/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "simplestake" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgBond:
			return handleMsgBond(ctx, k, msg)
		case MsgUnbond:
			return handleMsgUnbond(ctx, k, msg)
		default:
			return sdk.ErrUnknownRequest("No match for message type.").Result()
		}
	}
}

func handleMsgBond(ctx sdk.Context, k Keeper, msg MsgBond) sdk.Result {
	power, err := k.Bond(ctx, msg.Address, msg.PubKey, msg.Stake)
	if err != nil {
		return err.Result()
	}

	valSet := abci.Validator{
		PubKey: tmtypes.TM2PB.PubKey(msg.PubKey),
		Power:  power,
	}

	return sdk.Result{
		Code:             sdk.ABCICodeOK,
		ValidatorUpdates: abci.Validators{valSet},
	}
}

func handleMsgUnbond(ctx sdk.Context, k Keeper, msg MsgUnbond) sdk.Result {
	pubKey, _, err := k.Unbond(ctx, msg.Address)
	if err != nil {
		return err.Result()
	}

	valSet := abci.Validator{
		PubKey: tmtypes.TM2PB.PubKey(pubKey),
		Power:  int64(0),
	}

	return sdk.Result{
		Code:             sdk.ABCICodeOK,
		ValidatorUpdates: abci.Validators{valSet},
	}
}
