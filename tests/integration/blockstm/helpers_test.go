package blockstm_test

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/baseapp/txnrunner"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type blockExecutor interface {
	LastBlockHeight() int64
	LastCommitID() storetypes.CommitID
	FinalizeBlock(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)
	Commit() (*abci.ResponseCommit, error)
}

func newTestSTMRunner(txDecoder sdk.TxDecoder, storeKeys []storetypes.StoreKey, workers int) *txnrunner.STMRunner {
	return txnrunner.NewSTMRunner(
		txDecoder,
		storeKeys,
		workers,
		false,
		func(_ storetypes.MultiStore) string { return sdk.DefaultBondDenom },
	)
}

func finalizeNextBlock(t rapid.TB, app blockExecutor, txs [][]byte) *abci.ResponseFinalizeBlock {
	t.Helper()

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
		Txs:    txs,
	})
	require.NoError(t, err)

	return res
}

func commitBlock(t rapid.TB, app blockExecutor) storetypes.CommitID {
	t.Helper()

	_, err := app.Commit()
	require.NoError(t, err)

	return app.LastCommitID()
}

func finalizeAndCommitNextBlock(t rapid.TB, app blockExecutor, txs [][]byte) (*abci.ResponseFinalizeBlock, storetypes.CommitID) {
	t.Helper()

	res := finalizeNextBlock(t, app, txs)
	commitID := commitBlock(t, app)

	return res, commitID
}

func requireEquivalentBlockOutcome(
	t rapid.TB,
	expectedRes, actualRes *abci.ResponseFinalizeBlock,
	expectedCommitID, actualCommitID storetypes.CommitID,
) {
	t.Helper()

	require.Equal(t, expectedRes.TxResults, actualRes.TxResults)
	require.Equal(t, expectedRes.AppHash, actualRes.AppHash)
	require.Equal(t, expectedCommitID, actualCommitID)
}
