package types_test

import (
	"testing"

	"gotest.tools/v3/assert"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func TestAppGenesis_Marshal(t *testing.T) {
	genesis := types.AppGenesis{
		AppGenesisOnly: types.AppGenesisOnly{
			AppName:    "simapp",
			AppVersion: "0.1.0",
		},
		GenesisDoc: cmttypes.GenesisDoc{
			ChainID: "test",
		},
	}

	out, err := genesis.MarshalJSON()
	assert.NilError(t, err)

	assert.Equal(t, string(out), `{"app_name":"simapp","app_version":"0.1.0","chain_id":"test"}`)
}

func TestAppGenesis_Unmarshal(t *testing.T) {
	var genesis types.AppGenesis

	err := genesis.UnmarshalJSON([]byte(`{"app_name":"simapp","app_version":"0.1.0","chain_id":"test"}`))
	assert.NilError(t, err)

	assert.Equal(t, genesis.AppName, "simapp")
	assert.Equal(t, genesis.AppVersion, "0.1.0")
	assert.Equal(t, genesis.ChainID, "test")
}
