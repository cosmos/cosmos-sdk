package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

type runMsgsTxHandler struct {
	legacyRouter     sdk.Router        // router for redirecting legacy Msgs
	msgServiceRouter *MsgServiceRouter // router for redirecting Msg service messages
}

func NewRunMsgsTxHandler(msr *MsgServiceRouter, legacyRouter sdk.Router) tx.Handler {
	return runMsgsTxHandler{
		legacyRouter:     legacyRouter,
		msgServiceRouter: msr,
	}
}

var _ tx.Handler = runMsgsTxHandler{}

// CheckTx implements tx.Handler.CheckTx method.
func (txh runMsgsTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	// Don't run Msgs during CheckTx.
	return abci.ResponseCheckTx{}, nil
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (txh runMsgsTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	res, err := txh.runMsgs(sdk.UnwrapSDKContext(ctx), tx.GetMsgs(), req.Tx)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return abci.ResponseDeliverTx{
		// GasInfo will be populated by the Gas middleware.
		Log:    res.Log,
		Data:   res.Data,
		Events: res.Events,
	}, nil
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh runMsgsTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	res, err := txh.runMsgs(sdk.UnwrapSDKContext(ctx), sdkTx.GetMsgs(), req.TxBytes)
	if err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return tx.ResponseSimulateTx{
		// GasInfo will be populated by the Gas middleware.
		Result: res,
	}, nil
}

// runMsgs iterates through a list of messages and executes them with the provided
// Context and execution mode. Messages will only be executed during simulation
// and DeliverTx. An error is returned if any single message fails or if a
// Handler does not exist for a given message route. Otherwise, a reference to a
// Result is returned. The caller must not commit state if an error is returned.
func (txh runMsgsTxHandler) runMsgs(sdkCtx sdk.Context, msgs []sdk.Msg, txBytes []byte) (*sdk.Result, error) {
	// Create a new Context based off of the existing Context with a MultiStore branch
	// in case message processing fails. At this point, the MultiStore
	// is a branch of a branch.
	runMsgCtx, msCache := cacheTxContext(sdkCtx, txBytes)

	// Attempt to execute all messages and only update state if all messages pass
	// and we're in DeliverTx. Note, runMsgs will never return a reference to a
	// Result if any single message fails or does not have a registered Handler.
	msgLogs := make(sdk.ABCIMessageLogs, 0, len(msgs))
	events := sdkCtx.EventManager().Events()
	txMsgData := &sdk.TxMsgData{
		Data: make([]*sdk.MsgData, 0, len(msgs)),
	}

	// NOTE: GasWanted is determined by the Gas TxHandler and GasUsed by the GasMeter.
	for i, msg := range msgs {
		var (
			msgResult    *sdk.Result
			eventMsgName string // name to use as value in event `message.action`
			err          error
		)

		if handler := txh.msgServiceRouter.Handler(msg); handler != nil {
			// ADR 031 request type routing
			msgResult, err = handler(runMsgCtx, msg)
			eventMsgName = sdk.MsgTypeURL(msg)
		} else if legacyMsg, ok := msg.(legacytx.LegacyMsg); ok {
			// legacy sdk.Msg routing
			// Assuming that the app developer has migrated all their Msgs to
			// proto messages and has registered all `Msg services`, then this
			// path should never be called, because all those Msgs should be
			// registered within the `MsgServiceRouter` already.
			msgRoute := legacyMsg.Route()
			eventMsgName = legacyMsg.Type()
			handler := txh.legacyRouter.Route(sdkCtx, msgRoute)
			if handler == nil {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s; message index: %d", msgRoute, i)
			}

			msgResult, err = handler(sdkCtx, msg)
		} else {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "can't route message %+v", msg)
		}

		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to execute message; message index: %d", i)
		}

		msgEvents := sdk.Events{
			sdk.NewEvent(sdk.EventTypeMessage, sdk.NewAttribute(sdk.AttributeKeyAction, eventMsgName)),
		}
		msgEvents = msgEvents.AppendEvents(msgResult.GetEvents())

		// append message events, data and logs
		//
		// Note: Each message result's data must be length-prefixed in order to
		// separate each result.
		events = events.AppendEvents(msgEvents)

		txMsgData.Data = append(txMsgData.Data, &sdk.MsgData{MsgType: sdk.MsgTypeURL(msg), Data: msgResult.Data})
		msgLogs = append(msgLogs, sdk.NewABCIMessageLog(uint32(i), msgResult.Log, msgEvents))
	}

	msCache.Write()
	data, err := proto.Marshal(txMsgData)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to marshal tx data")
	}

	return &sdk.Result{
		Data:   data,
		Log:    strings.TrimSpace(msgLogs.String()),
		Events: events.ToABCIEvents(),
	}, nil
}

// cacheTxContext returns a new context based off of the provided context with
// a branched multi-store.
func cacheTxContext(sdkCtx sdk.Context, txBytes []byte) (sdk.Context, sdk.CacheMultiStore) {
	ms := sdkCtx.MultiStore()
	// TODO: https://github.com/cosmos/cosmos-sdk/issues/2824
	msCache := ms.CacheMultiStore()
	if msCache.TracingEnabled() {
		msCache = msCache.SetTracingContext(
			sdk.TraceContext(
				map[string]interface{}{
					"txHash": fmt.Sprintf("%X", tmhash.Sum(txBytes)),
				},
			),
		).(sdk.CacheMultiStore)
	}

	return sdkCtx.WithMultiStore(msCache), msCache
}
