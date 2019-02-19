package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ModuleName = "crisis"

func NewHandler(invars []InvarRoute) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {

		switch msg := msg.(type) {
		case MsgVerifyInvariance:
			return handleMsgVerifyInvariance(ctx, msg, invarianceRoutes)
		default:
			return sdk.ErrTxDecode("invalid message parse in crisis module").Result()
		}
	}
}

func handleMsgVerifyInvariance(ctx sdk.Context, msg MsgVerifyInvariance, invars []InvarRoute) sdk.Result {

	route := msg.InvarianceRoute
	found := false
	for _, invar := range invar {
		if invar.Route == msg.InvarianceRoute {
			invar.Invariant()
			found = true
		}
	}
	if !found {
		return ErrUnknownInvariant(DefaultCodespace).Result()
	}

	// if an invariant was broken, then reward sender account with their gas
	// cost plus a bonus from the community pool.
	txBytes, err := txBldr.BuildTxForSim(msgs)
	if err != nil {
		return
	}
	estimated, adjusted, err = CalculateGas(cliCtx.Query, cliCtx.Codec, txBytes, txBldr.GasAdjustment())

	tags := sdk.NewTags(
		"invariant", msg.InvarianceRoute,
	)
	return sdk.Result{
		Tags: tags,
	}
}
