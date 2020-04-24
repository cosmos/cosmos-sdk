package keys

import (
	"testing"

	"github.com/otiai10/copy"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/libs/cli"
)

func Test_runMigrateCmd(t *testing.T) {
	cmd := AddKeyCommand()
	assert.NotNil(t, cmd)

	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	copy.Copy("testdata", kbHome)
	assert.NotNil(t, kbHome)
	t.Cleanup(kbCleanUp)
	viper.Set(flags.FlagHome, kbHome)

	viper.Set(cli.OutputFlag, OutputFormatText)

	assert.NoError(t, runAddCmd(cmd, []string{"keyname1"}))

	viper.Set(flags.FlagDryRun, true)
	cmd = MigrateCommand()
	mockIn, _, _ := tests.ApplyMockIO(cmd)
	mockIn.Reset("test1234\ntest1234\n")
	assert.NoError(t, runMigrateCmd(cmd, []string{}))
}
