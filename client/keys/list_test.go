package keys

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runListCmd(t *testing.T) {
	cmd := ListKeysCmd()
	cmd.Flags().AddFlagSet(Commands().PersistentFlags())

	kbHome1, cleanUp1 := testutil.NewTestCaseDir(t)
	t.Cleanup(cleanUp1)

	kbHome2, cleanUp2 := testutil.NewTestCaseDir(t)
	t.Cleanup(cleanUp2)

	mockIn, _, _ := testutil.ApplyMockIO(cmd)
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome2, mockIn)
	require.NoError(t, err)

	path := "" //sdk.GetConfig().GetFullFundraiserPath()
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
				fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
			})

			if err := cmd.Execute(); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}

			cmd.SetArgs([]string{
				fmt.Sprintf("--%s=%s", flags.FlagHome, tt.kbDir),
				fmt.Sprintf("--%s=true", flagListNames),
				fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
			})

			if err := cmd.Execute(); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
