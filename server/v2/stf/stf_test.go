package stf

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/event"
	coregas "cosmossdk.io/core/gas"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/gas"
	"cosmossdk.io/server/v2/stf/mock"
)

const senderAddr = "sender"

func addMsgHandlerToSTF[T any, PT interface {
	*T
	transaction.Msg
},
	U any, UT interface {
		*U
		transaction.Msg
	}](
	t *testing.T,
	stf *STF[mock.Tx],
	handler func(ctx context.Context, msg PT) (UT, error),
) {
	t.Helper()
	msgRouterBuilder := NewMsgRouterBuilder()
	err := msgRouterBuilder.RegisterHandler(
		msgTypeURL(PT(new(T))),
		func(ctx context.Context, msg transaction.Msg) (msgResp transaction.Msg, err error) {
			typedReq := msg.(PT)
			typedResp, err := handler(ctx, typedReq)
			if err != nil {
				return nil, err
			}

			return typedResp, nil
		},
	)
	if err != nil {
		t.Errorf("Failed to register handler: %v", err)
	}

	msgRouter, err := msgRouterBuilder.build()
	if err != nil {
		t.Errorf("Failed to build message router: %v", err)
	}
	stf.msgRouter = msgRouter
}

func TestSTF(t *testing.T) {
	state := mock.DB()
	mockTx := mock.Tx{
		Sender:   []byte(senderAddr),
		Msg:      &gogotypes.BoolValue{Value: true},
		GasLimit: 100_000,
	}

	sum := sha256.Sum256([]byte("test-hash"))

	s := &STF[mock.Tx]{
		doPreBlock: func(ctx context.Context, txs []mock.Tx) error {
			ctx.(*executionContext).events = append(ctx.(*executionContext).events, event.NewEvent("pre-block"))
			return nil
		},
		doBeginBlock: func(ctx context.Context) error {
			kvSet(t, ctx, "begin-block")
			ctx.(*executionContext).events = append(ctx.(*executionContext).events, event.NewEvent("begin-block"))
			return nil
		},
		doEndBlock: func(ctx context.Context) error {
			kvSet(t, ctx, "end-block")
			ctx.(*executionContext).events = append(ctx.(*executionContext).events, event.NewEvent("end-block"))
			return nil
		},
		doValidatorUpdate: func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error) {
			ctx.(*executionContext).events = append(ctx.(*executionContext).events, event.NewEvent("validator-update"))
			return nil, nil
		},
		doTxValidation: func(ctx context.Context, tx mock.Tx) error {
			kvSet(t, ctx, "validate")
			ctx.(*executionContext).events = append(
				ctx.(*executionContext).events,
				event.NewEvent("validate-tx", event.NewAttribute(senderAddr, string(tx.Sender))),
				event.NewEvent(
					"validate-tx",
					event.NewAttribute(senderAddr, string(tx.Sender)),
					event.NewAttribute("index", "2"),
				),
			)
			return nil
		},
		postTxExec: func(ctx context.Context, tx mock.Tx, success bool) error {
			kvSet(t, ctx, "post-tx-exec")
			ctx.(*executionContext).events = append(
				ctx.(*executionContext).events,
				event.NewEvent("post-tx-exec", event.NewAttribute(senderAddr, string(tx.Sender))),
				event.NewEvent(
					"post-tx-exec",
					event.NewAttribute(senderAddr, string(tx.Sender)),
					event.NewAttribute("index", "2"),
				),
			)
			return nil
		},
		branchFn:            branch.DefaultNewWriterMap,
		makeGasMeter:        gas.DefaultGasMeter,
		makeGasMeteredState: gas.DefaultWrapWithGasMeter,
	}

	addMsgHandlerToSTF(t, s, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
		kvSet(t, ctx, "exec")
		ctx.(*executionContext).events = append(
			ctx.(*executionContext).events,
			event.NewEvent("handle-msg", event.NewAttribute("msg", msg.String())),
			event.NewEvent(
				"handle-msg",
				event.NewAttribute("msg", msg.String()),
				event.NewAttribute("index", "2"),
			),
		)
		return nil, nil
	})

	t.Run("begin and end block", func(t *testing.T) {
		_, newState, err := s.DeliverBlock(context.Background(), &server.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
		}, state)
		if err != nil {
			t.Errorf("DeliverBlock error: %v", err)
		}
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
	})

	t.Run("basic tx", func(t *testing.T) {
		result, newState, err := s.DeliverBlock(context.Background(), &server.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		if err != nil {
			t.Errorf("DeliverBlock error: %v", err)
		}
		stateHas(t, newState, "validate")
		stateHas(t, newState, "exec")
		stateHas(t, newState, "post-tx-exec")

		if len(result.TxResults) != 1 {
			t.Errorf("Expected 1 TxResult, got %d", len(result.TxResults))
		}
		txResult := result.TxResults[0]
		if txResult.GasUsed == 0 {
			t.Errorf("GasUsed should not be zero")
		}
		if txResult.GasWanted != mockTx.GasLimit {
			t.Errorf("Expected GasWanted to be %d, got %d", mockTx.GasLimit, txResult.GasWanted)
		}

		// Check PreBlockEvents
		preBlockEvents := result.PreBlockEvents
		if len(preBlockEvents) != 1 {
			t.Fatalf("Expected 1 PreBlockEvent, got %d", len(preBlockEvents))
		}
		if preBlockEvents[0].Type != "pre-block" {
			t.Errorf("Expected PreBlockEvent Type 'pre-block', got %s", preBlockEvents[0].Type)
		}
		if preBlockEvents[0].BlockStage != appdata.PreBlockStage {
			t.Errorf("Expected PreBlockStage %d, got %d", appdata.PreBlockStage, preBlockEvents[0].BlockStage)
		}
		if preBlockEvents[0].EventIndex != 1 {
			t.Errorf("Expected PreBlockEventIndex 1, got %d", preBlockEvents[0].EventIndex)
		}
		// Check BeginBlockEvents
		beginBlockEvents := result.BeginBlockEvents
		if len(beginBlockEvents) != 1 {
			t.Fatalf("Expected 1 BeginBlockEvent, got %d", len(beginBlockEvents))
		}
		if beginBlockEvents[0].Type != "begin-block" {
			t.Errorf("Expected BeginBlockEvent Type 'begin-block', got %s", beginBlockEvents[0].Type)
		}
		if beginBlockEvents[0].BlockStage != appdata.BeginBlockStage {
			t.Errorf("Expected BeginBlockStage %d, got %d", appdata.BeginBlockStage, beginBlockEvents[0].BlockStage)
		}
		if beginBlockEvents[0].EventIndex != 1 {
			t.Errorf("Expected BeginBlockEventIndex 1, got %d", beginBlockEvents[0].EventIndex)
		}
		// Check EndBlockEvents
		endBlockEvents := result.EndBlockEvents
		if len(endBlockEvents) != 2 {
			t.Fatalf("Expected 2 EndBlockEvents, got %d", len(endBlockEvents))
		}
		if endBlockEvents[0].Type != "end-block" {
			t.Errorf("Expected EndBlockEvent Type 'end-block', got %s", endBlockEvents[0].Type)
		}
		if endBlockEvents[1].Type != "validator-update" {
			t.Errorf("Expected EndBlockEvent Type 'validator-update', got %s", endBlockEvents[1].Type)
		}
		if endBlockEvents[1].BlockStage != appdata.EndBlockStage {
			t.Errorf("Expected EndBlockStage %d, got %d", appdata.EndBlockStage, endBlockEvents[1].BlockStage)
		}
		if endBlockEvents[0].EventIndex != 1 {
			t.Errorf("Expected EndBlockEventIndex 1, got %d", endBlockEvents[0].EventIndex)
		}
		if endBlockEvents[1].EventIndex != 2 {
			t.Errorf("Expected EndBlockEventIndex 2, got %d", endBlockEvents[1].EventIndex)
		}
		// check TxEvents
		events := txResult.Events
		if len(events) != 7 {
			t.Fatalf("Expected 7 TxEvents, got %d", len(events))
		}

		const message = "message"
		for i, event := range events {
			if event.BlockStage != appdata.TxProcessingStage {
				t.Errorf("Expected BlockStage %d, got %d", appdata.TxProcessingStage, event.BlockStage)
			}
			if event.TxIndex != 1 {
				t.Errorf("Expected TxIndex 1, got %d", event.TxIndex)
			}
			if event.EventIndex != int32(i%2+1) &&
				(event.Type == message && event.EventIndex != 3) { // special case for message event type as it happens in the msg handling flow
				t.Errorf("Expected EventIndex %d, got %d", i%2+1, event.EventIndex)
			}

			attrs, err := event.Attributes()
			if err != nil {
				t.Fatalf("Error getting event attributes: %v", err)
			}
			if len(attrs) < 1 || len(attrs) > 2 {
				t.Errorf("Expected 1 or 2 attributes, got %d", len(attrs))
			}

			if len(attrs) == 2 && event.Type != message {
				if attrs[1].Key != "index" || attrs[1].Value != "2" {
					t.Errorf("Expected attribute key 'index' and value '2', got key '%s' and value '%s'", attrs[1].Key, attrs[1].Value)
				}
			}
			switch i {
			case 0, 1:
				if event.Type != "validate-tx" {
					t.Errorf("Expected event type 'validate-tx', got %s", event.Type)
				}
				if event.MsgIndex != 0 {
					t.Errorf("Expected MsgIndex 0, got %d", event.MsgIndex)
				}
				if attrs[0].Key != senderAddr || attrs[0].Value != senderAddr {
					t.Errorf("Expected sender attribute key 'sender' and value 'sender', got key '%s' and value '%s'", attrs[0].Key, attrs[0].Value)
				}
			case 2, 3:
				if event.Type != "handle-msg" {
					t.Errorf("Expected event type 'handle-msg', got %s", event.Type)
				}
				if event.MsgIndex != 1 {
					t.Errorf("Expected MsgIndex 1, got %d", event.MsgIndex)
				}
				if attrs[0].Key != "msg" || attrs[0].Value != "&BoolValue{Value:true,XXX_unrecognized:[],}" {
					t.Errorf("Expected msg attribute with value '&BoolValue{Value:true,XXX_unrecognized:[],}', got '%s'", attrs[0].Value)
				}
			case 4:
				if event.Type != message {
					t.Errorf("Expected event type 'message', got %s", event.Type)
				}

				if event.MsgIndex != 1 {
					t.Errorf("Expected MsgIndex 1, got %d", event.MsgIndex)
				}

				if attrs[0].Key != "action" || attrs[0].Value != "/google.protobuf.BoolValue" {
					t.Errorf("Expected msg attribute with value '/google.protobuf.BoolValue', got '%s'", attrs[0].Value)
				}
			case 5, 6:
				if event.Type != "post-tx-exec" {
					t.Errorf("Expected event type 'post-tx-exec', got %s", event.Type)
				}
				if event.MsgIndex != -1 {
					t.Errorf("Expected MsgIndex -1, got %d", event.MsgIndex)
				}
				if attrs[0].Key != senderAddr || attrs[0].Value != senderAddr {
					t.Errorf("Expected sender attribute key 'sender' and value 'sender', got key '%s' and value '%s'", attrs[0].Key, attrs[0].Value)
				}
			}
		}
	})

	t.Run("exec tx out of gas", func(t *testing.T) {
		s := s.clone()

		mockTx := mock.Tx{
			Sender:   []byte(senderAddr),
			Msg:      &gogotypes.BoolValue{Value: true}, // msg does not matter at all because our handler does nothing.
			GasLimit: 0,                                 // NO GAS!
		}

		// this handler will propagate the storage error back, we expect
		// out of gas immediately at tx validation level.
		s.doTxValidation = func(ctx context.Context, tx mock.Tx) error {
			w, err := ctx.(*executionContext).state.GetWriter(actorName)
			if err != nil {
				t.Errorf("GetWriter error: %v", err)
			}
			err = w.Set([]byte("gas_failure"), []byte{})
			if err == nil {
				t.Error("Expected error, got nil")
			}
			return err
		}

		result, newState, err := s.DeliverBlock(context.Background(), &server.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		if err != nil {
			t.Errorf("DeliverBlock error: %v", err)
		}
		stateNotHas(t, newState, "gas_failure") // assert during out of gas no state changes leaked.
		if !errors.Is(result.TxResults[0].Error, coregas.ErrOutOfGas) {
			t.Errorf("Expected ErrOutOfGas, got %v", result.TxResults[0].Error)
		}
	})

	t.Run("fail exec tx", func(t *testing.T) {
		// update the stf to fail on the handler
		s := s.clone()
		addMsgHandlerToSTF(t, &s, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
			return nil, errors.New("failure")
		})

		blockResult, newState, err := s.DeliverBlock(context.Background(), &server.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		if err != nil {
			t.Errorf("DeliverBlock error: %v", err)
		}
		if !strings.Contains(blockResult.TxResults[0].Error.Error(), "failure") {
			t.Errorf("Expected error to contain 'failure', got %v", blockResult.TxResults[0].Error)
		}
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
		stateHas(t, newState, "validate")
		stateNotHas(t, newState, "exec")
		stateHas(t, newState, "post-tx-exec")
	})

	t.Run("tx is success but post tx failed", func(t *testing.T) {
		s := s.clone()
		s.postTxExec = func(ctx context.Context, tx mock.Tx, success bool) error {
			return errors.New("post tx failure")
		}
		blockResult, newState, err := s.DeliverBlock(context.Background(), &server.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		if err != nil {
			t.Errorf("DeliverBlock error: %v", err)
		}
		if !strings.Contains(blockResult.TxResults[0].Error.Error(), "post tx failure") {
			t.Errorf("Expected error to contain 'post tx failure', got %v", blockResult.TxResults[0].Error)
		}
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
		stateHas(t, newState, "validate")
		stateNotHas(t, newState, "exec")
		stateNotHas(t, newState, "post-tx-exec")
	})

	t.Run("tx failed and post tx failed", func(t *testing.T) {
		s := s.clone()
		addMsgHandlerToSTF(t, &s, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
			return nil, errors.New("exec failure")
		})
		s.postTxExec = func(ctx context.Context, tx mock.Tx, success bool) error { return errors.New("post tx failure") }
		blockResult, newState, err := s.DeliverBlock(context.Background(), &server.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		if err != nil {
			t.Errorf("DeliverBlock error: %v", err)
		}
		if !strings.Contains(blockResult.TxResults[0].Error.Error(), "exec failure\npost tx failure") {
			t.Errorf("Expected error to contain 'exec failure\npost tx failure', got %v", blockResult.TxResults[0].Error)
		}
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
		stateHas(t, newState, "validate")
		stateNotHas(t, newState, "exec")
		stateNotHas(t, newState, "post-tx-exec")
	})

	t.Run("fail validate tx", func(t *testing.T) {
		// update stf to fail on the validation step
		s := s.clone()
		s.doTxValidation = func(ctx context.Context, tx mock.Tx) error { return errors.New("failure") }
		blockResult, newState, err := s.DeliverBlock(context.Background(), &server.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		if err != nil {
			t.Errorf("DeliverBlock error: %v", err)
		}
		if !strings.Contains(blockResult.TxResults[0].Error.Error(), "failure") {
			t.Errorf("Expected error to contain 'failure', got %v", blockResult.TxResults[0].Error)
		}
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
		stateNotHas(t, newState, "validate")
		stateNotHas(t, newState, "exec")
	})

	t.Run("test validate tx with exec mode", func(t *testing.T) {
		// update stf to fail on the validation step
		s := s.clone()
		s.doTxValidation = func(ctx context.Context, tx mock.Tx) error {
			if ctx.(*executionContext).execMode == transaction.ExecModeCheck {
				return errors.New("failure")
			}
			return nil
		}
		// test ValidateTx as it validates with check execMode
		res := s.ValidateTx(context.Background(), state, mockTx.GasLimit, mockTx)
		if res.Error == nil {
			t.Error("Expected error, got nil")
		}

		// test validate tx with exec mode as finalize
		_, _, err := s.validateTx(context.Background(), s.branchFn(state), mockTx.GasLimit,
			mockTx, transaction.ExecModeFinalize)
		if err != nil {
			t.Errorf("validateTx error: %v", err)
		}
	})
}

var actorName = []byte("cookies")

func kvSet(t *testing.T, ctx context.Context, v string) {
	t.Helper()
	state, err := ctx.(*executionContext).state.GetWriter(actorName)
	if err != nil {
		t.Errorf("Set error: %v", err)
	} else {
		err = state.Set([]byte(v), []byte(v))
		if err != nil {
			t.Errorf("Set error: %v", err)
		}
	}
}

func stateHas(t *testing.T, accountState store.ReaderMap, key string) {
	t.Helper()
	state, err := accountState.GetReader(actorName)
	if err != nil {
		t.Errorf("GetReader error: %v", err)
	}
	has, err := state.Has([]byte(key))
	if err != nil {
		t.Errorf("Has error: %v", err)
	}
	if !has {
		t.Errorf("State did not have key: %s", key)
	}
}

func stateNotHas(t *testing.T, accountState store.ReaderMap, key string) {
	t.Helper()
	state, err := accountState.GetReader(actorName)
	if err != nil {
		t.Errorf("GetReader error: %v", err)
	}
	has, err := state.Has([]byte(key))
	if err != nil {
		t.Errorf("Has error: %v", err)
	}
	if has {
		t.Errorf("State was not supposed to have key: %s", key)
	}
}
