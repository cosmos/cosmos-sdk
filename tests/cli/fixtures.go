package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/params"
)

// Fixtures is used to setup the testing environment
type Fixtures struct {
	BuildDir       string
	RootDir        string
	SimdBinary     string
	ChainID        string
	RPCAddr        string
	Port           string
	SimdHome       string
	SimcliHome     string
	P2PAddr        string
	GRPCAddr     string
	Cdc            *codec.Codec
	EncodingConfig params.EncodingConfig
	T              *testing.T
}

// NewFixtures creates a new instance of Fixtures with many vars set
func NewFixtures(t *testing.T) *Fixtures {
	tmpDir, err := ioutil.TempDir("", "sdk_integration_"+t.Name()+"_")
	require.NoError(t, err)

	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)

	p2pAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)

	grpcAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)

	buildDir := os.Getenv("BUILDDIR")
	if buildDir == "" {
		t.Skip("builddir is empty, skipping")
	}

	encodingConfig := simapp.MakeEncodingConfig()

	return &Fixtures{
		T:              t,
		BuildDir:       buildDir,
		RootDir:        tmpDir,
		SimdBinary:     filepath.Join(buildDir, "simd"),
		SimdHome:       filepath.Join(tmpDir, ".simd"),
		SimcliHome:     filepath.Join(tmpDir, ".simcli"),
		RPCAddr:        servAddr,
		P2PAddr:        p2pAddr,
		GRPCAddr:     grpcAddr,
		Cdc:            encodingConfig.Amino,
		EncodingConfig: encodingConfig,
		Port:           port,
	}
}

// GenesisFile returns the path of the genesis file
func (f Fixtures) GenesisFile() string {
	return filepath.Join(f.SimdHome, "config", "genesis.json")
}

// GenesisFile returns the application's genesis state
func (f Fixtures) GenesisState() simapp.GenesisState {
	genDoc, err := tmtypes.GenesisDocFromFile(f.GenesisFile())
	require.NoError(f.T, err)

	var appState simapp.GenesisState
	require.NoError(f.T, f.Cdc.UnmarshalJSON(genDoc.AppState, &appState))
	return appState
}
