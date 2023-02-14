package types_test

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var expectedAppGenesisJSON = `{"app_name":"simapp","app_version":"0.1.0","genesis_time":"0001-01-01T00:00:00Z","chain_id":"test","initial_height":"0","app_hash":""}`

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

	out, err := json.Marshal(&genesis)
	assert.NilError(t, err)

	assert.Equal(t, string(out), expectedAppGenesisJSON)
}

func TestAppGenesis_Unmarshal(t *testing.T) {
	var genesis types.AppGenesis
	err := json.Unmarshal([]byte(expectedAppGenesisJSON), &genesis)
	assert.NilError(t, err)

	assert.DeepEqual(t, genesis.AppName, "simapp")
	assert.DeepEqual(t, genesis.AppVersion, "0.1.0")
	assert.DeepEqual(t, genesis.ChainID, "test")
}
