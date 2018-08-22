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
	democoindHome = ""
)

func init() {
	democoindHome = getTestingHomeDir()
}

func TestInitStartSequence(t *testing.T) {
	os.RemoveAll(democoindHome)
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
	out := tests.ExecuteT(t, fmt.Sprintf("democoind --home=%s init", democoindHome), "")
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(t, err)
	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(t, err)
}

func executeStart(t *testing.T, servAddr, port string) {
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("democoind start --home=%s --rpc.laddr=%v", democoindHome, servAddr))
	defer proc.Stop(false)
	tests.WaitForTMStart(port)
}

func getTestingHomeDir() string {
	tmpDir := os.TempDir()
	democoindHome := fmt.Sprintf("%s%s.test_democoind", tmpDir, string(os.PathSeparator))
	return democoindHome
}
