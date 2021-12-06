package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/crypto/tmhash"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
func (txh runMsgsTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	// Don't run Msgs during CheckTx.
	return tx.Response{}, tx.ResponseCheckTx{}, nil
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (txh runMsgsTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return txh.runMsgs(sdk.UnwrapSDKContext(ctx), req.Tx.GetMsgs(), req.TxBytes)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh runMsgsTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return txh.runMsgs(sdk.UnwrapSDKContext(ctx), req.Tx.GetMsgs(), req.TxBytes)
}

// runMsgs iterates through a list of messages and executes them with the provided
// Context and execution mode. Messages will only be executed during simulation
// and DeliverTx. An error is returned if any single message fails or if a
// Handler does not exist for a given message route. Otherwise, a reference to a
// Result is returned. The caller must not commit state if an error is returned.
func (txh runMsgsTxHandler) runMsgs(sdkCtx sdk.Context, msgs []sdk.Msg, txBytes []byte) (tx.Response, error) {
	// Create a new Context based off of the existing Context with a MultiStore branch
	// in case message processing fails. At this point, the MultiStore
	// is a branch of a branch.
	runMsgCtx, msCache := cacheTxContext(sdkCtx, txBytes)

	// Attempt to execute all messages and only update state if all messages pass
	// and we're in DeliverTx. Note, runMsgs will never return a reference to a
	// Result if any single message fails or does not have a registered Handler.
	msgLogs := make(sdk.ABCIMessageLogs, 0, len(msgs))
	events := sdkCtx.EventManager().Events()
	msgResponses := make([]*codectypes.Any, len(msgs))

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
				return tx.Response{}, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s; message index: %d", msgRoute, i)
			}

			msgResult, err = handler(sdkCtx, msg)
		} else {
			return tx.Response{}, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "can't route message %+v", msg)
		}

		if err != nil {
			return tx.Response{}, sdkerrors.Wrapf(err, "failed to execute message; message index: %d", i)
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

		// Each individual sdk.Result has exactly one Msg response. We aggregate here.
		msgResponse := msgResult.MsgResponses[0]
		if msgResponse == nil {
			return tx.Response{}, sdkerrors.ErrLogic.Wrapf("got nil Msg response at index %d for msg %s", i, sdk.MsgTypeURL(msg))
		}
		msgResponses[i] = msgResponse
		msgLogs = append(msgLogs, sdk.NewABCIMessageLog(uint32(i), msgResult.Log, msgEvents))
	}

	msCache.Write()

	return tx.Response{
		// GasInfo will be populated by the Gas middleware.
		Log:          strings.TrimSpace(msgLogs.String()),
		Events:       events.ToABCIEvents(),
		MsgResponses: msgResponses,
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
