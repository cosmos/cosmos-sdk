package keys

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runListCmd(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	cmdBasic := ListKeysCmd()

	// Prepare some keybases
	kbHome1, cleanUp1 := tests.NewTestCaseDir(t)
	t.Cleanup(cleanUp1)
	// Do nothing, leave home1 empty

	kbHome2, cleanUp2 := tests.NewTestCaseDir(t)
	t.Cleanup(cleanUp2)
	viper.Set(flags.FlagHome, kbHome2)

	mockIn, _, _ := tests.ApplyMockIO(cmdBasic)
	kb, err := keyring.New(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), mockIn)
	require.NoError(t, err)

	path := "" //sdk.GetConfig().GetFullFundraiserPath()
	_, err = kb.NewAccount("something", tests.TestMnemonic, "", path, hd.Secp256k1)
	require.NoError(t, err)

	t.Cleanup(func() {
		kb.Delete("something") // nolint:errcheck

	})
	testData := []struct {
		name    string
		kbDir   string
		args    args
		wantErr bool
	}{
		{"keybase: empty", kbHome1, args{cmdBasic, []string{}}, false},
		{"keybase: w/key", kbHome2, args{cmdBasic, []string{}}, false},
	}
	for _, tt := range testData {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(flagListNames, false)
			viper.Set(flags.FlagHome, tt.kbDir)
			if err := runListCmd(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}

			viper.Set(flagListNames, true)
			if err := runListCmd(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
