package keys

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttestutil "github.com/cosmos/cosmos-sdk/client/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func cleanupKeys(t *testing.T, kb keyring.Keyring, keys ...string) func() {
	return func() {
		for _, k := range keys {
			if err := kb.Delete(k); err != nil {
				t.Log("can't delete KB key ", k, err)
			}
		}
	}
}

func Test_runListCmd(t *testing.T) {
	cmd := ListKeysCmd()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

	kbHome1 := t.TempDir()
	kbHome2 := t.TempDir()

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	cdc := clienttestutil.MakeTestCodec(t)
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome2, mockIn, cdc)
	require.NoError(t, err)

	clientCtx := client.Context{}.WithKeyring(kb)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	path := "" // sdk.GetConfig().GetFullBIP44Path()
	_, err = kb.NewAccount("something", testdata.TestMnemonic, "", path, hd.Secp256k1)
	require.NoError(t, err)

	t.Cleanup(cleanupKeys(t, kb, "something"))

	testData := []struct {
		name    string
		kbDir   string
		wantErr bool
	}{
		{"keybase: empty", kbHome1, false},
		{"keybase: w/key", kbHome2, false},
	}
	for _, tt := range testData {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cmd.SetArgs([]string{
				fmt.Sprintf("--%s=%s", flags.FlagHome, tt.kbDir),
				fmt.Sprintf("--%s=false", flagListNames),
			})

			if err := cmd.ExecuteContext(ctx); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}

			cmd.SetArgs([]string{
				fmt.Sprintf("--%s=%s", flags.FlagHome, tt.kbDir),
				fmt.Sprintf("--%s=true", flagListNames),
			})

			if err := cmd.ExecuteContext(ctx); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
