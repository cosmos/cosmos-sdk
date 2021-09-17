package keys

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/stretchr/testify/require"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
)

func Test_runMigrateCmd(t *testing.T) {
	kbHome := t.TempDir()
	clientCtx := client.Context{}.WithKeyringDir(kbHome)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	require.NoError(t, copy.Copy("testdata", kbHome))

	cmd := MigrateCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	//mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	mockIn, mockOut := testutil.ApplyMockIO(cmd)

	cmd.SetArgs([]string{
		kbHome,
		//fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=true", flags.FlagDryRun),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	mockIn.Reset("\n12345678\n\n\n\n\n")
	t.Log(mockOut.String())
	assert.NoError(t, cmd.ExecuteContext(ctx))
}
