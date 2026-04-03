package simulation

import (
	"sync"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/tx/signing"
)

type testTxConfig struct{}

func (testTxConfig) TxEncoder() sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		return []byte(tx.(*testTx).id), nil
	}
}

func (testTxConfig) TxDecoder() sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		return &testTx{id: string(txBytes)}, nil
	}
}

func (testTxConfig) TxJSONEncoder() sdk.TxEncoder { return nil }

func (testTxConfig) TxJSONDecoder() sdk.TxDecoder { return nil }

func (testTxConfig) MarshalSignatureJSON(_ []signingtypes.SignatureV2) ([]byte, error) {
	return nil, nil
}

func (testTxConfig) UnmarshalSignatureJSON(_ []byte) ([]signingtypes.SignatureV2, error) {
	return nil, nil
}

func (testTxConfig) NewTxBuilder() client.TxBuilder {
	return nil
}

func (testTxConfig) WrapTxBuilder(_ sdk.Tx) (client.TxBuilder, error) {
	return nil, nil
}

func (testTxConfig) SignModeHandler() *signing.HandlerMap {
	return nil
}

func (testTxConfig) SigningContext() *signing.Context {
	return nil
}

var _ client.TxConfig = testTxConfig{}

type testTx struct {
	id string
}

func (*testTx) GetMsgs() []sdk.Msg { return nil }
func (*testTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

type testLifecycleApp struct {
	prepareResp *abci.ResponsePrepareProposal
	processResp *abci.ResponseProcessProposal
	deliveredTx *testTx
	prepareReq  *abci.RequestPrepareProposal
	processReq  *abci.RequestProcessProposal
}

func (a *testLifecycleApp) CheckTx(_ *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return &abci.ResponseCheckTx{Code: 0}, nil
}

func (a *testLifecycleApp) PrepareProposal(req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	a.prepareReq = req
	return a.prepareResp, nil
}

func (a *testLifecycleApp) ProcessProposal(req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	a.processReq = req
	return a.processResp, nil
}

func (a *testLifecycleApp) SimDeliver(_ sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	a.deliveredTx = tx.(*testTx)
	return sdk.GasInfo{}, &sdk.Result{}, nil
}

func TestExecuteTxLifecycle_UsesPreparedTxForDelivery(t *testing.T) {
	t.Parallel()

	app := &testLifecycleApp{
		prepareResp: &abci.ResponsePrepareProposal{
			Txs: [][]byte{[]byte("other"), []byte("original")},
		},
		processResp: &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT},
	}
	txConfig := testTxConfig{}
	inputTx := &testTx{id: "original"}
	ctx := sdk.Context{}.WithBlockHeight(10).WithBlockTime(time.Unix(1700000000, 0))

	outcome := ExecuteTxLifecycle(app, txConfig, inputTx, ctx)
	require.True(t, outcome.Accepted)
	require.NotNil(t, app.prepareReq)
	require.Equal(t, [][]byte{[]byte("original")}, app.prepareReq.Txs)
	require.NotNil(t, app.processReq)
	require.Equal(t, app.prepareResp.Txs, app.processReq.Txs)
	require.NotNil(t, app.deliveredTx)
	require.Equal(t, "original", app.deliveredTx.id)
}

func TestExecuteTxLifecycle_RejectsWhenProposalOmitsSimulatedTx(t *testing.T) {
	t.Parallel()

	app := &testLifecycleApp{
		prepareResp: &abci.ResponsePrepareProposal{
			Txs: [][]byte{[]byte("different")},
		},
		processResp: &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT},
	}
	txConfig := testTxConfig{}
	inputTx := &testTx{id: "original"}

	outcome := ExecuteTxLifecycle(app, txConfig, inputTx, sdk.Context{})
	require.False(t, outcome.Accepted)
	require.Equal(t, TxPhasePrepare, outcome.Phase)
	require.Equal(t, "proposal omitted simulated tx", outcome.Reason)
}

func TestTxLifecycleFailuresSnapshotForApp_IsolatedAcrossApps(t *testing.T) {
	t.Parallel()

	type keyedApp struct {
		id int
	}
	appA := &keyedApp{id: 1}
	appB := &keyedApp{id: 2}
	ResetTxLifecycleFailuresForApp(appA)
	ResetTxLifecycleFailuresForApp(appB)

	const n = 100
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			RecordTxLifecycleFailureForMsgForApp(appA, TxPhaseCheckTx, "msg/A", "reason A")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			RecordTxLifecycleFailureForMsgForApp(appB, TxPhaseFinalize, "msg/B", "reason B")
		}
	}()
	wg.Wait()

	snapshotA := TxLifecycleFailuresSnapshotForApp(appA)
	snapshotB := TxLifecycleFailuresSnapshotForApp(appB)

	require.Equal(t, n, snapshotA.TotalRejected)
	require.Equal(t, n, snapshotA.Phases[TxPhaseCheckTx].Total)
	require.Equal(t, 0, snapshotA.Phases[TxPhaseFinalize].Total)

	require.Equal(t, n, snapshotB.TotalRejected)
	require.Equal(t, n, snapshotB.Phases[TxPhaseFinalize].Total)
	require.Equal(t, 0, snapshotB.Phases[TxPhaseCheckTx].Total)
}
