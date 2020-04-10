package bank

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	defaultGenExportExpected = `{"send_enabled":true}`
	sendEnabledExpected      = false
)

var (
	genExportExpected = fmt.Sprintf(`{"send_enabled":%v}`, sendEnabledExpected)
)

func TestInitGenesis(t *testing.T) {
	input := setupTestInput()
	ctx, accKeeper, bankKeeper := input.ctx, input.ak, input.bk
	appModule := NewAppModule(bankKeeper, accKeeper)

	// 1.check default export
	require.Equal(t, defaultGenExportExpected, string(appModule.ExportGenesis(ctx)))

	// 2.change context
	bankKeeper.SetSendEnabled(ctx, sendEnabledExpected)

	// 3.export again
	genExport := appModule.ExportGenesis(ctx)
	require.Equal(t, genExportExpected, string(genExport))

	// 4.init again && check
	newInput := setupTestInput()
	newCtx, newAccKeeper, newBankKeeper := newInput.ctx, newInput.ak, newInput.bk
	newAppModule := NewAppModule(newBankKeeper, newAccKeeper)
	newAppModule.InitGenesis(newCtx, genExport)
	require.Equal(t, sendEnabledExpected, newBankKeeper.GetSendEnabled(newCtx))
}
