package types_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	"gotest.tools/v3/assert"
)

var expectedAppGenesisJSON = `{"app_name":"simapp","app_version":"0.1.0","genesis_time":"0001-01-01T00:00:00Z","chain_id":"test","initial_height":"0","app_hash":""}`

func TestAppGenesis_Marshal(t *testing.T) {
	genesis := types.AppGenesis{
		AppName:    "simapp",
		AppVersion: "0.1.0",
		ChainID:    "test",
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

func TestAppGenesis_ToCometBFTGenesisDoc(t *testing.T) {
	genesis := types.AppGenesis{
		AppName:    "simapp",
		AppVersion: "0.1.0",
		ChainID:    "test",
		AppHash:    []byte{5, 34, 11, 3, 23},
	}

	cmtGenesis, err := genesis.ToCometBFTGenesisDoc()
	assert.NilError(t, err)
	assert.DeepEqual(t, cmtGenesis.ChainID, "test")
	assert.DeepEqual(t, cmtGenesis.AppHash.String(), "05220B0317")
}
