package simsx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestMsgFactories(t *testing.T) {
	myMsg := testdata.NewTestMsg()
	mySigners := []SimAccount{{}}
	specs := map[string]struct {
		src           SimMsgFactoryX
		expErrHandled bool
	}{
		"default": {
			src: SimMsgFactoryFn[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg *testdata.TestMsg) {
				return mySigners, myMsg
			}),
		},
		"with delivery result handler": {
			src: NewSimMsgFactoryWithDeliveryResultHandler[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg *testdata.TestMsg, handler SimDeliveryResultHandler) {
				return mySigners, myMsg, func(err error) error { return nil }
			}),
			expErrHandled: true,
		},
		"with future ops": {
			src: NewSimMsgFactoryWithFutureOps[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter, fOpsReg FutureOpsRegistry) ([]SimAccount, *testdata.TestMsg) {
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

func TestLazyResultHandlers(t *testing.T) {
	myErr := errors.New("testing")
	passThrough := func(err error) error { return err }
	ignore := func(err error) error { return nil }
	specs := map[string]struct {
		handlers []SimDeliveryResultHandler
		expErrs  []error
	}{
		"handle err": {
			handlers: []SimDeliveryResultHandler{ignore},
			expErrs:  []error{nil},
		},
		"not handled": {
			handlers: []SimDeliveryResultHandler{passThrough},
			expErrs:  []error{myErr},
		},
		"nil value": {
			handlers: []SimDeliveryResultHandler{nil},
			expErrs:  []error{myErr},
		},
		"multiple": {
			handlers: []SimDeliveryResultHandler{ignore, passThrough, nil, passThrough},
			expErrs:  []error{nil, myErr, myErr, myErr},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var hPos int
			f := NewSimMsgFactoryWithDeliveryResultHandler[*testdata.TestMsg](func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg *testdata.TestMsg, handler SimDeliveryResultHandler) {
				defer func() { hPos++ }()
				return nil, nil, spec.handlers[hPos]
			})
			for i := 0; i < len(spec.handlers); i++ {
				_, _ = f.Create()(context.Background(), nil, nil)
				gotErr := f.DeliveryResultHandler()(myErr)
				assert.Equal(t, spec.expErrs[i], gotErr)
			}
		})
	}
}
