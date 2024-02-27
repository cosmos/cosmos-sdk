package baseapp

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
)

type TxExecutor func(
	ctx context.Context,
	blockSize int,
	cms storetypes.MultiStore,
	deliverTxWithMultiStore func(int, storetypes.MultiStore) *abci.ExecTxResult,
) ([]*abci.ExecTxResult, error)
