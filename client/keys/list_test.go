package keys

import (
	"context"
	"fmt"
	"testing"
	"errors"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runListCmd(t *testing.T) {
	cmd := ListKeysCmd()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

	kbHome1 := t.TempDir()
	kbHome2 := t.TempDir()

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	encCfg := simapp.MakeTestEncodingConfig()
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome2, mockIn, encCfg.Codec)
	require.NoError(t, err)

	clientCtx := client.Context{}.WithKeyring(kb)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	path := "" //sdk.GetConfig().GetFullBIP44Path()
	_, err = kb.NewAccount("something", testutil.TestMnemonic, "", path, hd.Secp256k1)
	require.NoError(t, err)

	t.Cleanup(func() {
		kb.Delete("something") // nolint:errcheck
	})

	type args struct {
		cmd  *cobra.Command
		args []string
	}

	testData := []struct {
		name    string
		kbDir   string
		wantErr error
	}{
		{"keybase: empty", kbHome1, keyring.ErrNoKeysAvailable},
		{"keybase: w/key", kbHome2, keyring.ErrNoKeysAvailable},
	}
	for _, tt := range testData {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cmd.SetArgs([]string{
				fmt.Sprintf("--%s=%s", flags.FlagHome, tt.kbDir),
				fmt.Sprintf("--%s=false", flagListNames),
				fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
			})

			if err := cmd.ExecuteContext(ctx); err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}

			cmd.SetArgs([]string{
				fmt.Sprintf("--%s=%s", flags.FlagHome, tt.kbDir),
				fmt.Sprintf("--%s=true", flagListNames),
				fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
			})

			if err := cmd.ExecuteContext(ctx); err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
