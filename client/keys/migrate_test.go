package keys

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/cli"
)

func Test_runMigrateCmd(t *testing.T) {

	cmdAddKey := addKeyCommand()
	assert.NotNil(t, cmdAddKey)
	mockIn, _, _ := tests.ApplyMockIO(cmdAddKey)

	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)

	viper.Set(cli.OutputFlag, OutputFormatText)

	mockIn.Reset("test1234\ntest1234\n")
	err := runAddCmd(cmdAddKey, []string{"keyname1"})
	assert.NoError(t, err)

	viper.Set(cli.OutputFlag, OutputFormatText)

	cmd := migrateKeyCommand()
	assert.NotNil(t, cmd)

	mockIn.Reset("test1234\n")

	err = runMigrateCmd(cmd, []string{""})
	assert.NoError(t, err)

}
