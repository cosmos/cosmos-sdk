package keys

import (
	"fmt"
	"testing"

	"github.com/99designs/keyring"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
)

func Test_runAddCmdBasic(t *testing.T) {

	backends := keyring.AvailableBackends()

	runningOnServer := false
	if len(backends) == 1 && backends[0] == keyring.BackendType("file") {
		runningOnServer = true
	}

	kb := NewKeyringKeybase()
	defer func() {
		kb.Delete("keyname1", "", false)
		kb.Delete("keyname2", "", false)

	}()

	cmd := addKeyCommand()
	assert.NotNil(t, cmd)
	mockIn, _, _ := tests.ApplyMockIO(cmd)

	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)

	viper.Set(cli.OutputFlag, OutputFormatText)

	if runningOnServer {
		fmt.Println("Running on Server")
		mockIn.Reset("testpass\ny\n")

	} else {
		mockIn.Reset("y\n")
	}
	err := runAddCmd(cmd, []string{"keyname1"})
	assert.NoError(t, err)

	viper.Set(cli.OutputFlag, OutputFormatText)

	mockIn.Reset("n\n")
	err = runAddCmd(cmd, []string{"keyname1"})
	assert.Error(t, err)

	viper.Set(cli.OutputFlag, OutputFormatText)

	mockIn.Reset("y\n")
	err = runAddCmd(cmd, []string{"keyname1"})
	assert.NoError(t, err)

	viper.Set(cli.OutputFlag, OutputFormatJSON)

	mockIn.Reset("y\n")
	err = runAddCmd(cmd, []string{"keyname2"})
	assert.NoError(t, err)

}
