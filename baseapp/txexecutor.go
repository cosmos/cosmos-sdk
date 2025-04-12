package baseapp

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxExecutor function type for implementing custom execution logic, such as block-stm
type TxExecutor func(
	ctx context.Context,
	block [][]byte,
	cms types.MultiStore,
	deliverTxWithMultiStore func(int, sdk.Tx, types.MultiStore, map[string]any) *abci.ExecTxResult,
) ([]*abci.ExecTxResult, error)
