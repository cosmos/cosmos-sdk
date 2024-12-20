package module

import (
	"context"
	"errors"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	common2 "github.com/cosmos/cosmos-sdk/simsx/runner/common"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgFactories(t *testing.T) {
	myMsg := testdata.NewTestMsg()
	mySigners := []common.SimAccount{{}}
	specs := map[string]struct {
		src           SimMsgFactoryX
		expErrHandled bool
	}{
		"default": {
			src: SimMsgFactoryFn[*testdata.TestMsg](func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) (signer []common.SimAccount, msg *testdata.TestMsg) {
				return mySigners, myMsg
			}),
		},
		"with delivery result handler": {
			src: NewSimMsgFactoryWithDeliveryResultHandler[*testdata.TestMsg](func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) (signer []common.SimAccount, msg *testdata.TestMsg, handler common.SimDeliveryResultHandler) {
				return mySigners, myMsg, func(err error) error { return nil }
			}),
			expErrHandled: true,
		},
		"with future ops": {
			src: NewSimMsgFactoryWithFutureOps[*testdata.TestMsg](func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter, fOpsReg common.FutureOpsRegistry) ([]common.SimAccount, *testdata.TestMsg) {
				return mySigners, myMsg
			}),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, (*testdata.TestMsg)(nil), spec.src.MsgType())

			factoryMethod := spec.src.Create()
			require.NotNil(t, factoryMethod)
			gotSigners, gotMsg := factoryMethod(context.Background(), nil, nil)
			assert.Equal(t, mySigners, gotSigners)
			assert.Equal(t, gotMsg, myMsg)

			require.NotNil(t, spec.src.DeliveryResultHandler())
			gotErr := spec.src.DeliveryResultHandler()(errors.New("testing"))
			assert.Equal(t, spec.expErrHandled, gotErr == nil)
		})
	}
}

func TestRunWithFailFast(t *testing.T) {
	myTestMsg := testdata.NewTestMsg()
	mySigners := []common.SimAccount{common.SimAccountFixture()}
	specs := map[string]struct {
		factory    FactoryMethod
		expSigners []common.SimAccount
		expMsg     sdk.Msg
		expSkipped bool
	}{
		"factory completes": {
			factory: func(ctx context.Context, _ *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, sdk.Msg) {
				return mySigners, myTestMsg
			},
			expSigners: mySigners,
			expMsg:     myTestMsg,
		},
		"factory skips": {
			factory: func(ctx context.Context, _ *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, sdk.Msg) {
				reporter.Skip("testing")
				return nil, nil
			},
			expSkipped: true,
		},
		"factory skips and panics": {
			factory: func(ctx context.Context, _ *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, sdk.Msg) {
				reporter.Skip("testing")
				panic("should be handled")
			},
			expSkipped: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, done := context.WithCancel(context.Background())
			reporter := common.NewBasicSimulationReporter().WithScope(&testdata.TestMsg{}, common.SkipHookFn(func(...any) { done() }))
			gotSigners, gotMsg := common2.SafeRunFactoryMethod(ctx, nil, reporter, spec.factory)
			assert.Equal(t, spec.expSigners, gotSigners)
			assert.Equal(t, spec.expMsg, gotMsg)
			assert.Equal(t, spec.expSkipped, reporter.IsAborted())
		})
	}
}
