package keys

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/stretchr/testify/require"
	"github.com/cosmos/cosmos-sdk/testutil"
)

func Test_runMigrateCmd(t *testing.T) {
	require := require.New(t)
	kbHome := t.TempDir()
	clientCtx := client.Context{}.WithKeyringDir(kbHome)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	//require.NoError(copy.Copy("testdata", kbHome))

	cmd := MigrateCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	//mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	mockIn, mockOut := testutil.ApplyMockIO(cmd)

	mockIn.Reset("\n12345678\n\n\n\n\n")
	t.Log(mockOut.String())
	require.NoError(cmd.ExecuteContext(ctx))
}
