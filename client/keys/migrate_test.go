package keys

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/libs/cli"
)

func Test_runMigrateCmd(t *testing.T) {
	config := sdk.NewDefaultConfig()
	cmd := AddKeyCommand(config)
	assert.NotNil(t, cmd)
	mockIn, _, _ := tests.ApplyMockIO(cmd)

	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)

	viper.Set(cli.OutputFlag, OutputFormatText)

	mockIn.Reset("test1234\ntest1234\n")
	err := runAddCmd(config)(cmd, []string{"keyname1"})
	assert.NoError(t, err)

	viper.Set(flags.FlagDryRun, true)
	cmd = MigrateCommand(config)
	mockIn, _, _ = tests.ApplyMockIO(cmd)
	mockIn.Reset("test1234\n")
	assert.NoError(t, runMigrateCmd(config)(cmd, []string{}))
}
