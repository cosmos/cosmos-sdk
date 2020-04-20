package helpers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// Fixtures is used to setup the testing environment
type Fixtures struct {
	BuildDir     string
	RootDir      string
	SimdBinary   string
	SimcliBinary string
	ChainID      string
	RPCAddr      string
	Port         string
	SimdHome     string
	SimcliHome   string
	P2PAddr      string
	T            *testing.T
}

// NewFixtures creates a new instance of Fixtures with many vars set
func NewFixtures(t *testing.T) *Fixtures {
	tmpDir, err := ioutil.TempDir("", "sdk_integration_"+t.Name()+"_")
	require.NoError(t, err)

	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)

	p2pAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)

	buildDir := os.Getenv("BUILDDIR")
	if buildDir == "" {
		buildDir, err = filepath.Abs("../../build/")
		require.NoError(t, err)
	}

	return &Fixtures{
		T:            t,
		BuildDir:     buildDir,
		RootDir:      tmpDir,
		SimdBinary:   filepath.Join(buildDir, "simd"),
		SimcliBinary: filepath.Join(buildDir, "simcli"),
		SimdHome:     filepath.Join(tmpDir, ".simd"),
		SimcliHome:   filepath.Join(tmpDir, ".simcli"),
		RPCAddr:      servAddr,
		P2PAddr:      p2pAddr,
		Port:         port,
	}
}

// GenesisFile returns the path of the genesis file
func (f Fixtures) GenesisFile() string {
	return filepath.Join(f.SimdHome, "config", "genesis.json")
}

// GenesisFile returns the application's genesis state
func (f Fixtures) GenesisState() simapp.GenesisState {
	cdc := codec.New()
	genDoc, err := tmtypes.GenesisDocFromFile(f.GenesisFile())
	require.NoError(f.T, err)

	var appState simapp.GenesisState
	require.NoError(f.T, cdc.UnmarshalJSON(genDoc.AppState, &appState))
	return appState
}
