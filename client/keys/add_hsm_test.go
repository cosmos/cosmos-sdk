package keys

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestAddCmdHsmKey(t *testing.T) {

	cmd := AddKeyCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

	// Prepare a keybase
	kbHome := t.TempDir()

	cdc := simapp.MakeTestEncodingConfig().Codec
	clientCtx := client.Context{}.WithKeyringDir(kbHome).WithCodec(cdc)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	userHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	DefaultNodeHome := filepath.Join(userHomeDir, ".simapp")

	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagHome, DefaultNodeHome),
		fmt.Sprintf("--%s=true", flags.FlagUseHsm),
		fmt.Sprintf("--%s=0", flagAccount),
		fmt.Sprintf("--%s=0", flagIndex),
		fmt.Sprintf("--%s=%s", cli.OutputFlag, OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyAlgorithm, string(hd.Secp256k1Type)),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	mockIn.Reset("test1234\ntest1234\n")
	require.NoError(t, cmd.ExecuteContext(ctx))

	// Now check that it has been stored properly
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)

	require.NoError(t, err)
	require.NotNil(t, kb)

	t.Cleanup(func() {
		_ = kb.Delete("keyname1")
	})

	mockIn.Reset("test1234\n")
	key1, err := kb.Key("keyname1")
	require.NoError(t, err)
	require.NotNil(t, key1)

	require.Equal(t, "keyname1", key1.Name)
	require.Equal(t, keyring.TypeHsm, key1.GetType())
	pub, err := key1.GetPubKey()
	require.NoError(t, err)
	fmt.Printf("Pubkey: %s", pub.String())
}
