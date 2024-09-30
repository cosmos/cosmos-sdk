package stf

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

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

	msgRouter, err := msgRouterBuilder.Build()
	if err != nil {
		t.Errorf("Failed to build message router: %v", err)
	}
	stf.msgRouter = msgRouter
}

func TestSTF(t *testing.T) {
	state := mock.DB()
	mockTx := mock.Tx{
		Sender:   []byte("sender"),
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
				event.NewEvent("validate-tx", event.NewAttribute("sender", string(tx.Sender))),
				event.NewEvent(
					"validate-tx",
					event.NewAttribute("sender", string(tx.Sender)),
					event.NewAttribute("index", "2"),
				),
			)
			return nil
		},
		postTxExec: func(ctx context.Context, tx mock.Tx, success bool) error {
			kvSet(t, ctx, "post-tx-exec")
			ctx.(*executionContext).events = append(
				ctx.(*executionContext).events,
				event.NewEvent("post-tx-exec", event.NewAttribute("sender", string(tx.Sender))),
				event.NewEvent(
					"post-tx-exec",
					event.NewAttribute("sender", string(tx.Sender)),
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

		// check PreBlockEvents
		require.Len(t, result.PreBlockEvents, 1)
		require.Equal(t, "pre-block", result.PreBlockEvents[0].Type)
		require.Equal(t, appdata.PreBlockStage, result.PreBlockEvents[0].BlockStage)
		require.Equal(t, int32(1), result.PreBlockEvents[0].EventIndex)
		// check BeginBlockEvents
		require.Len(t, result.BeginBlockEvents, 1)
		require.Equal(t, "begin-block", result.BeginBlockEvents[0].Type)
		require.Equal(t, appdata.BeginBlockStage, result.BeginBlockEvents[0].BlockStage)
		require.Equal(t, int32(1), result.BeginBlockEvents[0].EventIndex)
		// check EndBlockEvents
		require.Len(t, result.EndBlockEvents, 2)
		require.Equal(t, "end-block", result.EndBlockEvents[0].Type)
		require.Equal(t, "validator-update", result.EndBlockEvents[1].Type)
		require.Equal(t, appdata.EndBlockStage, result.EndBlockEvents[1].BlockStage)
		require.Equal(t, int32(1), result.EndBlockEvents[0].EventIndex)
		require.Equal(t, int32(2), result.EndBlockEvents[1].EventIndex)
		// check TxEvents
		require.Len(t, txResult.Events, 6)
		for i, event := range txResult.Events {
			require.Equal(t, appdata.TxProcessingStage, event.BlockStage)
			require.Equal(t, int32(1), event.TxIndex)
			require.Equal(t, int32(i%2+1), event.EventIndex)
			attrs, err := event.Attributes()
			require.NoError(t, err)
			require.Less(t, len(attrs), 3)
			require.Greater(t, len(attrs), 0)
			if len(attrs) == 2 {
				require.Equal(t, "index", attrs[1].Key)
				require.Equal(t, "2", attrs[1].Value)
			}
			switch i {
			case 0, 1:
				require.Equal(t, "validate-tx", event.Type)
				require.Equal(t, int32(0), event.MsgIndex)
				require.Equal(t, "sender", attrs[0].Key)
				require.Equal(t, "sender", attrs[0].Value)
			case 2, 3:
				require.Equal(t, "handle-msg", event.Type)
				require.Equal(t, int32(1), event.MsgIndex)
				require.Equal(t, "msg", attrs[0].Key)
				require.Equal(t, "&BoolValue{Value:true,XXX_unrecognized:[],}", attrs[0].Value)

			case 4, 5:
				require.Equal(t, "post-tx-exec", event.Type)
				require.Equal(t, int32(-1), event.MsgIndex)
				require.Equal(t, "sender", attrs[0].Key)
				require.Equal(t, "sender", attrs[0].Value)
			}
		}
	})

	t.Run("exec tx out of gas", func(t *testing.T) {
		s := s.clone()

		mockTx := mock.Tx{
			Sender:   []byte("sender"),
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
