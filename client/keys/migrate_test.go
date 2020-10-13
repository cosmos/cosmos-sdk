package keys

import (
	"context"
	"fmt"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
)

func Test_runMigrateCmd(t *testing.T) {
	cmd := AddKeyCommand()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

	kbHome := t.TempDir()

	clientCtx := client.Context{}.WithKeyringDir(kbHome)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	copy.Copy("testdata", kbHome)
	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=%s", cli.OutputFlag, OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	assert.NoError(t, cmd.ExecuteContext(ctx))

	cmd = MigrateCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

	cmd.SetArgs([]string{
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=true", flags.FlagDryRun),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	mockIn.Reset("test1234\ntest1234\n")
	assert.NoError(t, cmd.ExecuteContext(ctx))
}
