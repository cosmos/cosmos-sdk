package keys

import (
	"fmt"
	"testing"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/cli"
)

func Test_runMigrateCmd(t *testing.T) {
	backends := keyring.AvailableBackends()

	runningOnServer := false

	if len(backends) == 2 && backends[1] == keyring.BackendType("file") {
		runningOnServer = true
	}
	cmdAddKey := addKeyCommand()
	assert.NotNil(t, cmdAddKey)
	mockIn, _, _ := tests.ApplyMockIO(cmdAddKey)

	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)

	cmd := migrateKeyCommand()

	kb := NewKeyringKeybase(mockIn)

	defer func() {
		kb.Delete("keyname1", "", false)
	}()

	viper.Set(cli.OutputFlag, OutputFormatText)

	viper.Set(flags.FlagLegacy, true)

	mockIn.Reset("test1234\ntest1234\n")

	err := runAddCmd(cmdAddKey, []string{"keyname1"})
	assert.NoError(t, err)

	fmt.Println("Key Generated")

	viper.Set(cli.OutputFlag, OutputFormatText)

	assert.NotNil(t, cmd)
	mockIn, _, _ = tests.ApplyMockIO(cmd)

	if runningOnServer {
		mockIn.Reset("testpass1\ntestpass1\ntest1234\ntestpass1\n")
	} else {
		mockIn.Reset("test1234\n")
	}
	err = runMigrateCmd(cmd, []string{""})
	assert.NoError(t, err)

}
