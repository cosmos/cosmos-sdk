package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
)

func Test_runAddCmdBasic(t *testing.T) {
	cmd := addKeyCommand()
	assert.NotNil(t, cmd)
	mockIn, _, _ := tests.ApplyMockIO(cmd)

	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)

	viper.Set(cli.OutputFlag, OutputFormatText)

	mockIn.Reset("test1234\ntest1234\n")
	err := runAddCmd(cmd, []string{"keyname1"})
	assert.NoError(t, err)

	viper.Set(cli.OutputFlag, OutputFormatText)

	mockIn.Reset("test1234\ntest1234\n")
	err = runAddCmd(cmd, []string{"keyname1"})
	assert.Error(t, err)

	viper.Set(cli.OutputFlag, OutputFormatText)

	mockIn.Reset("y\ntest1234\ntest1234\n")
	err = runAddCmd(cmd, []string{"keyname1"})
	assert.NoError(t, err)

	viper.Set(cli.OutputFlag, OutputFormatJSON)

	mockIn.Reset("test1234\ntest1234\n")
	err = runAddCmd(cmd, []string{"keyname2"})
	assert.NoError(t, err)
}
