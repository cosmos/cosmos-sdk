package blockexec

import (
	"cmp"
	"fmt"
	goruntime "runtime"
	"slices"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/txnrunner"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Option overrides default behavior in Apply.
type Option func(*options)

type options struct {
	defaultExecutor    string
	defaultPreEstimate bool
	wrapRunner         func(sdk.TxRunner) sdk.TxRunner
}

// WithDefaultExecutor sets the executor used when appOpts has no block-executor
// value (e.g. programmatic construction); a bound flag/app.toml value still
// wins. Defaults to config.DefaultBlockExecutor (sequential).
func WithDefaultExecutor(executor string) Option {
	return func(o *options) { o.defaultExecutor = executor }
}

// WithDefaultPreEstimate sets the pre-estimate used when appOpts has no
// block-stm-pre-estimate value; a bound flag/app.toml value still wins.
func WithDefaultPreEstimate(v bool) Option {
	return func(o *options) { o.defaultPreEstimate = v }
}

// WithRunnerWrap wraps the TxRunner before installation — e.g. EVM chains
// use this so PatchTxResponses runs once per block regardless of executor.
func WithRunnerWrap(wrap func(sdk.TxRunner) sdk.TxRunner) Option {
	return func(o *options) { o.wrapRunner = wrap }
}

// Apply resolves the executor from appOpts (with Option overrides) and
// installs the corresponding TxRunner on bApp. Unknown executors panic.
func Apply(
	bApp *baseapp.BaseApp,
	appOpts servertypes.AppOptions,
	stores []storetypes.StoreKey,
	txDecoder sdk.TxDecoder,
	coinDenom func(storetypes.MultiStore) string,
	opts ...Option,
) {
	o := options{
		defaultExecutor: config.DefaultBlockExecutor,
	}
	for _, opt := range opts {
		opt(&o)
	}

	executor := cast.ToString(appOpts.Get(server.FlagBlockExecutor))
	if executor == "" {
		executor = o.defaultExecutor
	}

	var runner sdk.TxRunner
	switch executor {
	case config.BlockExecutorBlockSTM:
		workers := cast.ToInt(appOpts.Get(server.FlagBlockSTMWorkers))
		if workers <= 0 {
			workers = min(goruntime.GOMAXPROCS(0), goruntime.NumCPU())
		}

		preEstimate := o.defaultPreEstimate
		if v := appOpts.Get(server.FlagBlockSTMPreEstimate); v != nil {
			preEstimate = cast.ToBool(v)
		}

		sorted := slices.Clone(stores)
		slices.SortFunc(sorted, func(a, b storetypes.StoreKey) int {
			return cmp.Compare(a.Name(), b.Name())
		})

		bApp.Logger().Info("installing block-stm tx runner",
			"workers", workers, "pre_estimate", preEstimate, "wrapped", o.wrapRunner != nil)
		runner = txnrunner.NewSTMRunner(txDecoder, sorted, workers, preEstimate, coinDenom)

		// Disable the block gas meter before installing a parallel runner:
		// SetBlockSTMTxRunner panics if the meter is still enabled.
		bApp.SetDisableBlockGasMeter(true)

	case config.BlockExecutorSequential:
		bApp.Logger().Info("installing sequential tx runner", "wrapped", o.wrapRunner != nil)
		if o.wrapRunner == nil {
			// Leave BaseApp's lazy DefaultRunner in place.
			return
		}
		runner = txnrunner.NewDefaultRunner(txDecoder)

	default:
		panic(fmt.Errorf("unknown block executor: %s", executor))
	}

	if o.wrapRunner != nil {
		runner = o.wrapRunner(runner)
	}
	bApp.SetBlockSTMTxRunner(runner)
}
