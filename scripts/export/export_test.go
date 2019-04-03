package export

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	app "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	path        = "./genesis.json"
	chainID     = "cosmos-zone"
	genesisTime = "2019-02-11T12:00:00Z"
)

func TestNewGenesisFile(t *testing.T) {
	cdc := app.MakeCodec()
	genDoc, err := defaultGenesisDoc(chainID)
	require.NoError(t, err)

	output, err := cdc.MarshalJSONIndent(genDoc, "", " ")
	require.NoError(t, err)

	err = ioutil.WriteFile(path, output, 0644)
	require.NoError(t, err)

	genesisFile, err := NewGenesisFile(cdc, path)
	require.NoError(t, err)
	require.NotEqual(t, GenesisFile{}, genesisFile)
	os.Remove(path)
}

func TestDefaultGenesisDoc(t *testing.T) {
	expectedGenDoc := tmtypes.GenesisDoc{ChainID: chainID}
	genDoc, err := defaultGenesisDoc(chainID)
	require.NoError(t, err)
	require.NotEqual(t, expectedGenDoc, genDoc)

	genDoc, err = defaultGenesisDoc("")
	require.Error(t, err)
}

func TestImportGenesis(t *testing.T) {
	genesis := tmtypes.GenesisDoc{ChainID: chainID}

	output, err := json.Marshal(genesis)
	require.NoError(t, err)

	err = ioutil.WriteFile(path, output, 0644)
	require.NoError(t, err)

	genDoc, err := importGenesis(path)
	require.NoError(t, err)
	require.NotEqual(t, genesis, genDoc)
	os.Remove(path)

	// should fail with invalid genesis
	genesis = tmtypes.GenesisDoc{}
	output, err = json.Marshal(genesis)
	require.NoError(t, err)

	err = ioutil.WriteFile(path, output, 0644)
	require.NoError(t, err)

	genDoc, err = importGenesis(path)
	require.Error(t, err)
	os.Remove(path)
}
