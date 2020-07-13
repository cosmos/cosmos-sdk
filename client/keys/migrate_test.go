package keys

import (
	"fmt"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
)

func Test_runMigrateCmd(t *testing.T) {
	cmd := AddKeyCommand()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)
	cmd.Flags().AddFlagSet(Commands().PersistentFlags())

	kbHome, kbCleanUp := testutil.NewTestCaseDir(t)
	copy.Copy("testdata", kbHome)
	assert.NotNil(t, kbHome)
	t.Cleanup(kbCleanUp)

	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=%s", cli.OutputFlag, OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	assert.NoError(t, cmd.Execute())

	cmd = MigrateCommand()
	cmd.Flags().AddFlagSet(Commands().PersistentFlags())
	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

	cmd.SetArgs([]string{
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=true", flags.FlagDryRun),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	mockIn.Reset("test1234\ntest1234\n")
	assert.NoError(t, cmd.Execute())
}
