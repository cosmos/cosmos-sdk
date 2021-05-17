package keys

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runDeleteCmd(t *testing.T) {
	// Now add a temporary keybase
	kbHome := t.TempDir()
	cmd := DeleteKeyCommand()
	cmd.Flags().AddFlagSet(Commands(kbHome).PersistentFlags())
	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

	yesF, _ := cmd.Flags().GetBool(flagYes)
	forceF, _ := cmd.Flags().GetBool(flagForce)

	require.False(t, yesF)
	require.False(t, forceF)

	fakeKeyName1 := "runDeleteCmd_Key1"
	fakeKeyName2 := "runDeleteCmd_Key2"

	path := sdk.GetConfig().GetFullFundraiserPath()

	cmd.SetArgs([]string{"blah", fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome)})
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn)
	require.NoError(t, err)

	_, err = kb.NewAccount(fakeKeyName1, testutil.TestMnemonic, "", path, hd.Secp256k1)
	require.NoError(t, err)

	_, _, err = kb.NewMnemonic(fakeKeyName2, keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	require.NoError(t, err)

	clientCtx := client.Context{}.
		WithKeyringDir(kbHome).
		WithKeyring(kb)

	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	err = cmd.ExecuteContext(ctx)
	require.Error(t, err)
	require.Equal(t, "The specified item could not be found in the keyring", err.Error())

	// User confirmation missing
	cmd.SetArgs([]string{
		fakeKeyName1,
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	err = cmd.Execute()
	require.Error(t, err)
	require.Equal(t, "EOF", err.Error())

	_, err = kb.Key(fakeKeyName1)
	require.NoError(t, err)

	// Now there is a confirmation
	cmd.SetArgs([]string{
		fakeKeyName1,
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=true", flagYes),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.NoError(t, cmd.Execute())

	_, err = kb.Key(fakeKeyName1)
	require.Error(t, err) // Key1 is gone

	_, err = kb.Key(fakeKeyName2)
	require.NoError(t, err)

	cmd.SetArgs([]string{
		fakeKeyName2,
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=true", flagYes),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.NoError(t, cmd.Execute())

	_, err = kb.Key(fakeKeyName2)
	require.Error(t, err) // Key2 is gone
}
