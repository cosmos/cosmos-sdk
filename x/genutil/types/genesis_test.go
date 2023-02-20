package types_test

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var expectedAppGenesisJSON = `{"app_name":"simapp","app_version":"0.1.0","genesis_time":"0001-01-01T00:00:00Z","chain_id":"test","initial_height":0,"app_hash":null}`

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

func TestAppGenesis_ValidCometBFTGenesis(t *testing.T) {
	// validate can read cometbft genesis file
	genesis, err := types.AppGenesisFromFile("testdata/cmt_genesis.json")
	assert.NilError(t, err)

	assert.DeepEqual(t, genesis.ChainID, "demo")
	assert.DeepEqual(t, genesis.Validators[0].Name, "test")

	// validate the app genesis can be translated properly to cometbft genesis
	cmtGenesis, err := genesis.ToCometBFTGenesisDoc()
	assert.NilError(t, err)
	rawCmtGenesis, err := cmttypes.GenesisDocFromFile("testdata/cmt_genesis.json")
	assert.NilError(t, err)
	assert.DeepEqual(t, cmtGenesis, rawCmtGenesis)

	// validate can properly marshal to app genesis file
	rawAppGenesis, err := json.Marshal(&genesis)
	assert.NilError(t, err)
	golden.Assert(t, string(rawAppGenesis), "app_genesis.json")
}
