package baseapp

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RunTxModeDeliver() RunTxMode {
	return runTxModeDeliver
}

var (
	feeKeyBytes = []byte(sdk.AttributeKeyFee)
)

// getFeeFromTags get actual system_fee from Result
func getFeeFromTags(ctx sdk.Context, res sdk.Result) (eventI, attrI int, fee sdk.DecCoins) {
	if ctx.BlockHeight() < 1 {
		return -1, -1, sdk.DecCoins{}
	}
	for i, event := range res.Events {
		if event.Type == sdk.EventTypeMessage {
			for j, attr := range event.Attributes {
				if bytes.EqualFold(attr.GetKey(), feeKeyBytes) {
					if string(attr.Value) == "0"+sdk.DefaultBondDenom || string(attr.Value) == "0.00000000"+sdk.DefaultBondDenom {
						return i, j, sdk.DecCoins{}
					}
					fee, err := sdk.ParseDecCoins(string(attr.Value))
					if err != nil {
						panic(fmt.Sprintf("fee attribute's value is not valid, err=%s", err.Error()))
					}
					return i, j, fee
				}
			}
		}
	}

	//if fee is not set, e.g. MsgOrder/MsgCreateValidator...
	return -1, -1, sdk.DecCoins{}
}

func removeFeeTags(res sdk.Result, eventI, attrI int) sdk.Result {
	if eventI < 0 || attrI < 0 {
		return res
	}

	attrs := res.Events[eventI].Attributes
	if attrI < len(attrs)-1 {
		res.Events[eventI].Attributes = append(res.Events[eventI].Attributes, attrs[attrI+1:]...)
	} else {
		res.Events[eventI].Attributes = attrs[:attrI]
	}

	return res
}

//-------------------------------------------------------
// BaseApp
//-------------------------------------------------------

// Returns the applications's deliverState if app is in RunTxModeDeliver,
// otherwise it returns the application's checkstate.
func (app *BaseApp) GetState(mode RunTxMode) *State {
	if mode == runTxModeCheck || mode == runTxModeSimulate {
		return app.checkState
	}

	return app.deliverState
}

func (app *BaseApp) GetCommitMultiStore() sdk.CommitMultiStore {
	return app.cms
}

func (app *BaseApp) GetDeliverStateCtx() sdk.Context {
	return app.deliverState.ctx
}

//-------------------------------------------------------
// for protocol engine to invoke
//-------------------------------------------------------
func (app *BaseApp) PushInitChainer(initChainer sdk.InitChainer) {
	app.initChainer = initChainer
}

func (app *BaseApp) PushBeginBlocker(beginBlocker sdk.BeginBlocker) {
	app.beginBlocker = beginBlocker
}

func (app *BaseApp) PushEndBlocker(endBlocker sdk.EndBlocker) {
	app.endBlocker = endBlocker
}

func (app *BaseApp) PushAnteHandler(ah sdk.AnteHandler) {
	app.anteHandler = ah
}

func (app *BaseApp) SetTxDecoder(txDecoder sdk.TxDecoder) {
	app.txDecoder = txDecoder
}

func (app *BaseApp) SetRouter(router sdk.Router, queryRouter sdk.QueryRouter) {
	app.router = router
	app.queryRouter = queryRouter
}
