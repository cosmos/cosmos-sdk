package auth

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	defaultGenExportExpected       = `{"params":{"max_memo_characters":"256","tx_sig_limit":"7","tx_size_cost_per_byte":"10","sig_verify_cost_ed25519":"590","sig_verify_cost_secp256k1":"1000"}}`
	maxMemoCharactersExpected      = uint64(512)
	txSigLimitExpected             = uint64(14)
	txSizeCostPerByteExpected      = uint64(20)
	sigVerifyCostED25519Expected   = uint64(1180)
	sigVerifyCostSecp256k1Expected = uint64(2000)
)

var (
	genExportExpected = fmt.Sprintf(`{"params":{"max_memo_characters":"%d","tx_sig_limit":"%d","tx_size_cost_per_byte":"%d","sig_verify_cost_ed25519":"%d","sig_verify_cost_secp256k1":"%d"}}`,
		maxMemoCharactersExpected, txSigLimitExpected, txSizeCostPerByteExpected, sigVerifyCostED25519Expected, sigVerifyCostSecp256k1Expected)
)

func TestInitGenesis(t *testing.T) {
	input := setupTestInput()
	ctx, accKeeper := input.ctx, input.ak
	appModule := NewAppModule(accKeeper)

	// 1.check default export
	require.Equal(t, defaultGenExportExpected, string(appModule.ExportGenesis(ctx)))

	// 2.change context
	newParams := types.NewParams(maxMemoCharactersExpected, txSigLimitExpected, txSizeCostPerByteExpected,
		sigVerifyCostED25519Expected, sigVerifyCostSecp256k1Expected)
	accKeeper.SetParams(ctx, newParams)

	// 3.export again
	genExport := appModule.ExportGenesis(ctx)
	require.Equal(t, genExportExpected, string(genExport))

	// 4.init again && check
	newInput := setupTestInput()
	newCtx, newAccKeeper := newInput.ctx, newInput.ak
	newAppModule := NewAppModule(newAccKeeper)
	newAppModule.InitGenesis(newCtx, genExport)
	newParams = newAccKeeper.GetParams(newCtx)
	require.Equal(t, maxMemoCharactersExpected, newParams.MaxMemoCharacters)
	require.Equal(t, txSigLimitExpected, newParams.TxSigLimit)
	require.Equal(t, txSizeCostPerByteExpected, newParams.TxSizeCostPerByte)
	require.Equal(t, sigVerifyCostED25519Expected, newParams.SigVerifyCostED25519)
	require.Equal(t, sigVerifyCostSecp256k1Expected, newParams.SigVerifyCostSecp256k1)
}
