package simplestake

import (
	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "simplestake" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case BondMsg:
			return handleBondMsg(ctx, k, msg)
		case UnbondMsg:
			return handleUnbondMsg(ctx, k, msg)
		default:
			return sdk.ErrUnknownRequest("No match for message type.").Result()
		}
	}
}

func handleBondMsg(ctx sdk.Context, k Keeper, msg BondMsg) sdk.Result {
	power, err := k.Bond(ctx, msg.Address, msg.PubKey, msg.Stake)
	if err != nil {
		return err.Result()
	}

	valSet := abci.Validator{
		PubKey: msg.PubKey.Bytes(),
		Power:  power,
	}

	return sdk.Result{
		Code:             sdk.CodeOK,
		ValidatorUpdates: abci.Validators{valSet},
	}
}

func handleUnbondMsg(ctx sdk.Context, k Keeper, msg UnbondMsg) sdk.Result {
	pubKey, _, err := k.Unbond(ctx, msg.Address)
	if err != nil {
		return err.Result()
	}

	valSet := abci.Validator{
		PubKey: pubKey.Bytes(),
		Power:  int64(0),
	}

	return sdk.Result{
		Code:             sdk.CodeOK,
		ValidatorUpdates: abci.Validators{valSet},
	}
}
