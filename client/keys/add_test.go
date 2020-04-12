package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runAddCmdBasic(t *testing.T) {
	cmd := AddKeyCommand()
	require.NotNil(t, cmd)
	mockIn, _, _ := tests.ApplyMockIO(cmd)

	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	require.NotNil(t, kbHome)
	t.Cleanup(kbCleanUp)
	viper.Set(flags.FlagHome, kbHome)
	viper.Set(cli.OutputFlag, OutputFormatText)
	viper.Set(flags.FlagUseLedger, false)

	mockIn.Reset("y\n")
	kb, err := keyring.New(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), kbHome, mockIn)
	require.NoError(t, err)
	t.Cleanup(func() {
		kb.Delete("keyname1") // nolint:errcheck
		kb.Delete("keyname2") // nolint:errcheck
	})
	require.NoError(t, runAddCmd(cmd, []string{"keyname1"}))

	mockIn.Reset("N\n")
	require.Error(t, runAddCmd(cmd, []string{"keyname1"}))

	require.NoError(t, runAddCmd(cmd, []string{"keyname2"}))
	require.Error(t, runAddCmd(cmd, []string{"keyname2"}))
	mockIn.Reset("y\n")
	require.NoError(t, runAddCmd(cmd, []string{"keyname2"}))

	// test --dry-run
	require.NoError(t, runAddCmd(cmd, []string{"keyname4"}))
	require.Error(t, runAddCmd(cmd, []string{"keyname4"}))

	viper.Set(flags.FlagDryRun, true)
	require.NoError(t, runAddCmd(cmd, []string{"keyname4"}))
}
