package baseapp

import (
	"sync"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.AnteDecorator = lockAndCacheContextDecorator{}

func NewLockAndCacheContextAnteDecorator() sdk.AnteDecorator {
	return lockAndCacheContextDecorator{
		mtx: &sync.Mutex{},
	}
}

type lockAndCacheContextDecorator struct {
	mtx *sync.Mutex
}

func (l lockAndCacheContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	var cacheMs storetypes.CacheMultiStore
	ctx, cacheMs = cacheTxContext(ctx, ctx.TxBytes())
	newCtx, err := next(ctx, tx, simulate)
	if err == nil {
		cacheMs.Write()
	}
	return newCtx, err
}
