package clitest

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/stretchr/testify/require"
)

var (
	basecoindHome = ""
)

func init() {
	basecoindHome = getTestingHomeDir()
}

func TestInitStartSequence(t *testing.T) {
	os.RemoveAll(basecoindHome)
	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	executeInit(t)
	executeStart(t, servAddr, port)
}

func executeInit(t *testing.T) {
	var (
		chainID string
		initRes map[string]json.RawMessage
	)
	out := tests.ExecuteT(t, fmt.Sprintf("basecoind --home=%s init", basecoindHome), "")
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(t, err)
	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(t, err)
}

func executeStart(t *testing.T, servAddr, port string) {
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("basecoind start --home=%s --rpc.laddr=%v", basecoindHome, servAddr))
	defer proc.Stop(false)
	tests.WaitForTMStart(port)
}

func getTestingHomeDir() string {
	tmpDir := os.TempDir()
	basecoindHome := fmt.Sprintf("%s%s.test_basecoind", tmpDir, string(os.PathSeparator))
	return basecoindHome
}
