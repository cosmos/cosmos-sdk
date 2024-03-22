package baseapp

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/store/types"
)

type TxExecutor func(
	ctx context.Context,
	blockSize int,
	cms types.MultiStore,
	deliverTxWithMultiStore func(int, types.MultiStore) *abci.ExecTxResult,
) ([]*abci.ExecTxResult, error)
