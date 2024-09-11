package simsx

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestSimsMsgRegistryAdapter(t *testing.T) {
	senderAcc := SimAccountFixture()
	accs := []simtypes.Account{senderAcc.Account}
	ak := MockAccountSourceX{GetAccountFn: MemoryAccountSource(senderAcc).GetAccount}
	myMsg := testdata.NewTestMsg(senderAcc.Address)
	ctx := sdk.Context{}.WithContext(context.Background())
	futureTime := time.Now().Add(time.Second)

	specs := map[string]struct {
		factory           SimMsgFactoryX
		expFactoryMsg     sdk.Msg
		expFactoryErr     error
		expDeliveryErr    error
		expFutureOpsCount int
	}{
		"successful execution": {
			factory: SimMsgFactoryFn[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg *testdata.TestMsg) {
				return []SimAccount{senderAcc}, myMsg
			}),
			expFactoryMsg: myMsg,
		},
		"skip execution": {
			factory: SimMsgFactoryFn[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg *testdata.TestMsg) {
				reporter.Skip("testing")
				return nil, nil
			}),
		},
		"future ops registration": {
			factory: NewSimMsgFactoryWithFutureOps[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter, fOpsReg FutureOpsRegistry) (signer []SimAccount, msg *testdata.TestMsg) {
				fOpsReg.Add(futureTime, SimMsgFactoryFn[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg *testdata.TestMsg) {
					return []SimAccount{senderAcc}, myMsg
				}))
				return []SimAccount{senderAcc}, myMsg
			}),
			expFactoryMsg:     myMsg,
			expFutureOpsCount: 1,
		},
		"error in factory execution": {
			factory: NewSimMsgFactoryWithFutureOps[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter, fOpsReg FutureOpsRegistry) (signer []SimAccount, msg *testdata.TestMsg) {
				reporter.Fail(errors.New("testing"))
				return nil, nil
			}),
			expFactoryErr: errors.New("testing"),
		},
		"missing senders": {
			factory: SimMsgFactoryFn[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg *testdata.TestMsg) {
				return nil, myMsg
			}),
			expDeliveryErr: errors.New("no senders"),
		},
		"error in delivery execution": {
			factory: SimMsgFactoryFn[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg *testdata.TestMsg) {
				return []SimAccount{senderAcc}, myMsg
			}),
			expDeliveryErr: errors.New("testing"),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			r := NewBasicSimulationReporter()
			reg := NewSimsMsgRegistryAdapter(r, ak, nil, txConfig(), log.NewNopLogger())
			// when
			reg.Add(100, spec.factory)
			// then
			gotOps := reg.ToLegacyObjects()
			require.Len(t, gotOps, 1)
			assert.Equal(t, 100, gotOps[0].Weight())

			// and when ops executed
			var capturedTXs []sdk.Tx
			captureTXApp := AppEntrypointFn(func(_txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
				capturedTXs = append(capturedTXs, tx)
				return sdk.GasInfo{}, &sdk.Result{}, spec.expDeliveryErr
			})
			fn := gotOps[0].Op()
			gotOpsResult, gotFOps, gotErr := fn(rand.New(rand.NewSource(1)), captureTXApp, ctx, accs, "testchain")
			// then
			if spec.expFactoryErr != nil {
				require.Equal(t, spec.expFactoryErr, gotErr)
				assert.Empty(t, gotFOps)
				assert.Equal(t, gotOpsResult.OK, spec.expFactoryErr == nil)
				assert.Empty(t, gotOpsResult.Comment)
				require.Empty(t, capturedTXs)
			}
			if spec.expDeliveryErr != nil {
				require.Equal(t, spec.expDeliveryErr, gotErr)
			}
			// and verify TX delivery
			if spec.expFactoryMsg != nil {
				require.Len(t, capturedTXs, 1)
				require.Len(t, capturedTXs[0].GetMsgs(), 1)
				assert.Equal(t, spec.expFactoryMsg, capturedTXs[0].GetMsgs()[0])
			}
			assert.Len(t, gotFOps, spec.expFutureOpsCount)
		})
	}
}

func TestRunWithFailFast(t *testing.T) {
	myTestMsg := testdata.NewTestMsg()
	mySigners := []SimAccount{SimAccountFixture()}
	specs := map[string]struct {
		factory    FactoryMethod
		expSigners []SimAccount
		expMsg     sdk.Msg
		expSkipped bool
	}{
		"factory completes": {
			factory: func(ctx context.Context, _ *ChainDataSource, reporter SimulationReporter) ([]SimAccount, sdk.Msg) {
				return mySigners, myTestMsg
			},
			expSigners: mySigners,
			expMsg:     myTestMsg,
		},
		"factory skips": {
			factory: func(ctx context.Context, _ *ChainDataSource, reporter SimulationReporter) ([]SimAccount, sdk.Msg) {
				reporter.Skip("testing")
				return nil, nil
			},
			expSkipped: true,
		},
		"factory skips and panics": {
			factory: func(ctx context.Context, _ *ChainDataSource, reporter SimulationReporter) ([]SimAccount, sdk.Msg) {
				reporter.Skip("testing")
				panic("should be handled")
			},
			expSkipped: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, done := context.WithCancel(context.Background())
			reporter := NewBasicSimulationReporter().WithScope(&testdata.TestMsg{}, SkipHookFn(func(...any) { done() }))
			gotSigners, gotMsg := runWithFailFast(ctx, nil, reporter, spec.factory)
			assert.Equal(t, spec.expSigners, gotSigners)
			assert.Equal(t, spec.expMsg, gotMsg)
			assert.Equal(t, spec.expSkipped, reporter.IsSkipped())
		})
	}
}
