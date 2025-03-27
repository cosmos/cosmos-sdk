package protocolpool

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/testutil"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func TestCreateTestModule(t *testing.T) {
	app, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			testutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
	)
	assert.NilError(t, err)

	gotModule, found := app.ModuleManager.Modules[types.ModuleName]
	assert.Assert(t, found)
	assert.Assert(t, gotModule != nil)

	gotModule, found = app.ModuleManager.Modules[distrtypes.ModuleName]
	assert.Assert(t, found)
	assert.Assert(t, gotModule != nil)
}
