package v1

import (
	"errors"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestDeliverSimsMsg(t *testing.T) {
	var (
		sender   = common.SimAccountFixture()
		ak       = common.MemoryAccountSource(sender)
		myMsg    = testdata.NewTestMsg(sender.Address)
		txConfig = txConfig()
		r        = rand.New(rand.NewSource(1))
		ctx      sdk.Context
	)
	noopResultHandler := func(err error) error { return err }
	specs := map[string]struct {
		app                      AppEntrypoint
		reporter                 func() common.SimulationReporterRuntime
		deliveryResultHandler    common.SimDeliveryResultHandler
		errDeliveryResultHandler error
		expOps                   simtypes.OperationMsg
	}{
		"error when reporter skipped": {
			app: SimDeliverFn(func(_txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
				return sdk.GasInfo{GasWanted: 100, GasUsed: 20}, &sdk.Result{}, nil
			}),
			reporter: func() common.SimulationReporterRuntime {
				r := common.NewBasicSimulationReporter()
				r.Skip("testing")
				return r
			},
			expOps: simtypes.NoOpMsg("", "", "testing"),
		},
		"successful delivery": {
			app: SimDeliverFn(func(_txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
				return sdk.GasInfo{GasWanted: 100, GasUsed: 20}, &sdk.Result{}, nil
			}),
			reporter:              func() common.SimulationReporterRuntime { return common.NewBasicSimulationReporter() },
			deliveryResultHandler: noopResultHandler,
			expOps:                simtypes.NewOperationMsgBasic("", "", "", true),
		},
		"error delivery": {
			app: SimDeliverFn(func(_txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
				return sdk.GasInfo{GasWanted: 100, GasUsed: 20}, &sdk.Result{}, errors.New("my error")
			}),
			reporter:              func() common.SimulationReporterRuntime { return common.NewBasicSimulationReporter() },
			deliveryResultHandler: noopResultHandler,
			expOps:                simtypes.NewOperationMsgBasic("", "", "delivering tx with msgs: &testdata.TestMsg{Signers:[]string{\"cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3\"}, DecField:0.000000000000000000}", false),
		},
		"error delivery handled": {
			app: SimDeliverFn(func(_txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
				return sdk.GasInfo{GasWanted: 100, GasUsed: 20}, &sdk.Result{}, errors.New("my error")
			}),
			reporter:              func() common.SimulationReporterRuntime { return common.NewBasicSimulationReporter() },
			deliveryResultHandler: func(err error) error { return nil },
			expOps:                simtypes.NewOperationMsgBasic("", "", "", true),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := DeliverSimsMsg(ctx, spec.reporter(), spec.app, r, txConfig, ak, "testing", myMsg, spec.deliveryResultHandler, sender)
			assert.Equal(t, spec.expOps, got)
		})
	}
}

func TestSimulationReporterToLegacy(t *testing.T) {
	myErr := errors.New("my-error")
	myMsg := testdata.NewTestMsg([]byte{1})

	specs := map[string]struct {
		setup  func() common.SimulationReporterRuntime
		expOp  simtypes.OperationMsg
		expErr error
	}{
		"init only": {
			setup:  func() common.SimulationReporterRuntime { return common.NewBasicSimulationReporter() },
			expOp:  simtypes.NewOperationMsgBasic("", "", "", false),
			expErr: errors.New("operation aborted before msg was executed"),
		},
		"success result": {
			setup: func() common.SimulationReporterRuntime {
				r := common.NewBasicSimulationReporter().WithScope(myMsg)
				r.Success(myMsg, "testing")
				return r
			},
			expOp: simtypes.NewOperationMsgBasic("TestMsg", "/testpb.TestMsg", "testing", true),
		},
		"error result": {
			setup: func() common.SimulationReporterRuntime {
				r := common.NewBasicSimulationReporter().WithScope(myMsg)
				r.Fail(myErr, "testing")
				return r
			},
			expOp:  simtypes.NewOperationMsgBasic("TestMsg", "/testpb.TestMsg", "testing", false),
			expErr: myErr,
		},
		"last error wins": {
			setup: func() common.SimulationReporterRuntime {
				r := common.NewBasicSimulationReporter().WithScope(myMsg)
				r.Fail(errors.New("other-err"), "testing1")
				r.Fail(myErr, "testing2")
				return r
			},
			expOp:  simtypes.NewOperationMsgBasic("TestMsg", "/testpb.TestMsg", "testing1, testing2", false),
			expErr: myErr,
		},
		"skipped ": {
			setup: func() common.SimulationReporterRuntime {
				r := common.NewBasicSimulationReporter().WithScope(myMsg)
				r.Skip("testing")
				return r
			},
			expOp: simtypes.NoOpMsg("TestMsg", "/testpb.TestMsg", "testing"),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			r := spec.setup()
			assert.Equal(t, spec.expOp, toLegacyOperationMsg(r))
			require.Equal(t, spec.expErr, r.Close())
		})
	}
}
