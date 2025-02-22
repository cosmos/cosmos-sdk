//go:build system_test

package systemtests

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	systest "cosmossdk.io/systemtests"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func TestChainInit(t *testing.T) {
	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	removeGenesis(t)
	// init with height
	testInitialHeight := int64(333)
	cli.RunCommandWithArgs("init", "test-height", "--initial-height", fmt.Sprintf("%d", testInitialHeight), "--home="+systest.Sut.NodeDir(0))
	appGenesis, err := genutiltypes.AppGenesisFromFile(systest.Sut.NodeDir(0) + "/config/genesis.json")
	require.NoError(t, err)
	require.Equal(t, testInitialHeight, appGenesis.InitialHeight)

	removeGenesis(t)
	// init with negative height
	testInitialHeight = -333
	cli.RunCommandWithArgs("init", "test-height", "--initial-height", fmt.Sprintf("%d", testInitialHeight), "--home="+systest.Sut.NodeDir(0))
	appGenesis, err = genutiltypes.AppGenesisFromFile(systest.Sut.NodeDir(0) + "/config/genesis.json")
	require.NoError(t, err)
	require.Equal(t, int64(1), appGenesis.InitialHeight)

	removeGenesis(t)
	// init with custom denom
	customDenom := "mydenom"
	cli.RunCommandWithArgs("init", "test-denom", "--default-denom", customDenom, "--home="+systest.Sut.NodeDir(0))
	appGenesis, err = genutiltypes.AppGenesisFromFile(systest.Sut.NodeDir(0) + "/config/genesis.json")
	require.NoError(t, err)
	var appState struct {
		Staking struct {
			Params struct {
				BondDenom string `json:"bond_denom"`
			} `json:"params"`
		} `json:"staking"`
	}
	err = json.Unmarshal(appGenesis.AppState, &appState)
	require.NoError(t, err)
	require.Equal(t, customDenom, appState.Staking.Params.BondDenom)

	removeGenesis(t)
	// init with recover
	mnemonic := strings.NewReader("decide praise business actor peasant farm drastic weather extend front hurt later song give verb rhythm worry fun pond reform school tumble august one")
	cli.RunCommandWithInputAndArgs(mnemonic, "init", "test-denom", "--recover", "--home="+systest.Sut.NodeDir(0))
	_, err = os.Stat(path.Join(systest.Sut.NodeDir(0), "config", "genesis.json"))
	require.NoError(t, err)
}

func removeGenesis(t *testing.T) {
	require.NoError(t, os.Remove(path.Join(systest.Sut.NodeDir(0), "config", "genesis.json")))
}
