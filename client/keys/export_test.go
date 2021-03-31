package keys

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/testutil"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runExportCmd(t *testing.T) {
	cmd := ExportKeyCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

	// Now add a temporary keybase
	kbHome := t.TempDir()

	// create a key
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn)
	require.NoError(t, err)
	t.Cleanup(func() {
		kb.Delete("keyname1") // nolint:errcheck
	})

	path := sdk.GetConfig().GetFullFundraiserPath()
	_, err = kb.NewAccount("keyname1", testutil.TestMnemonic, "", path, hd.Secp256k1)
	require.NoError(t, err)

	// Now enter password
	args := []string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	}

	mockIn.Reset("123456789\n123456789\n")
	cmd.SetArgs(args)

	clientCtx := client.Context{}.
		WithKeyringDir(kbHome).
		WithKeyring(kb)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	require.NoError(t, cmd.ExecuteContext(ctx))

	argsUnsafeOnly := append(args, "--unsafe")
	cmd.SetArgs(argsUnsafeOnly)
	require.Error(t, cmd.ExecuteContext(ctx))

	argsUnarmoredHexOnly := append(args, "--unarmored-hex")
	cmd.SetArgs(argsUnarmoredHexOnly)
	require.Error(t, cmd.ExecuteContext(ctx))

	argsUnsafeUnarmoredHex := append(args, "--unsafe", "--unarmored-hex")
	cmd.SetArgs(argsUnsafeUnarmoredHex)
	require.Error(t, cmd.ExecuteContext(ctx))

	mockIn, mockOut := testutil.ApplyMockIO(cmd)
	mockIn.Reset("y\n")
	require.NoError(t, cmd.ExecuteContext(ctx))
	require.Equal(t, "2485e33678db4175dc0ecef2d6e1fc493d4a0d7f7ce83324b6ed70afe77f3485\n", mockOut.String())
}
