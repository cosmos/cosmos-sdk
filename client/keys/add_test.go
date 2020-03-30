package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runAddCmdBasic(t *testing.T) {
	cmd := AddKeyCommand()
	assert.NotNil(t, cmd)
	mockIn, _, _ := tests.ApplyMockIO(cmd)

	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	t.Cleanup(kbCleanUp)
	viper.Set(flags.FlagHome, kbHome)
	viper.Set(cli.OutputFlag, OutputFormatText)

	mockIn.Reset("y\n")
	kb, err := keyring.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), kbHome, mockIn)
	require.NoError(t, err)
	t.Cleanup(func() {
		kb.Delete("keyname1", "", false) // nolint:errcheck
		kb.Delete("keyname2", "", false) // nolint:errcheck
	})
	assert.NoError(t, runAddCmd(cmd, []string{"keyname1"}))

	mockIn.Reset("N\n")
	assert.Error(t, runAddCmd(cmd, []string{"keyname1"}))

	assert.NoError(t, runAddCmd(cmd, []string{"keyname2"}))
	assert.Error(t, runAddCmd(cmd, []string{"keyname2"}))
	mockIn.Reset("y\n")
	assert.NoError(t, runAddCmd(cmd, []string{"keyname2"}))

	// test --dry-run
	assert.NoError(t, runAddCmd(cmd, []string{"keyname4"}))
	assert.Error(t, runAddCmd(cmd, []string{"keyname4"}))

	viper.Set(flags.FlagDryRun, true)
	assert.NoError(t, runAddCmd(cmd, []string{"keyname4"}))
}
