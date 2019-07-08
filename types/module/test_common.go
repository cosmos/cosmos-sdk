package module

import (
	"fmt"
	"testing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// run the tx through the anteHandler and ensure its valid
func CheckValidTx(t *testing.T, m Manager, ctx sdk.Context, tx sdk.Tx, simulate bool) {
	_, result, abort := m.AnteHandler(ctx, tx, simulate)
	require.Equal(t, "", result.Log)
	require.False(t, abort)
	require.Equal(t, sdk.CodeOK, result.Code)
	require.True(t, result.IsOK())
}

// run the tx through the anteHandler and ensure it fails with the given code
func CheckInvalidTx(t *testing.T, m Manager, ctx sdk.Context, tx sdk.Tx, simulate bool, code sdk.CodeType) {
	newCtx, result, abort := m.AnteHandler(ctx, tx, simulate)
	require.True(t, abort)

	require.Equal(t, code, result.Code, fmt.Sprintf("Expected %v, got %v", code, result))
	require.Equal(t, sdk.CodespaceRoot, result.Codespace)

	if code == sdk.CodeOutOfGas {
		// GasWanted set correctly
		require.Equal(t, tx.Gas(), result.GasWanted, "Gas wanted not set correctly")
		require.True(t, result.GasUsed > result.GasWanted, "GasUsed not greated than GasWanted")
		// Check that context is set correctly
		require.Equal(t, result.GasUsed, newCtx.GasMeter().GasConsumed(), "Context not updated correctly")
	}
}
