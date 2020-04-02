package keys

import (
	"bufio"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runDeleteCmd(t *testing.T) {
	deleteKeyCommand := DeleteKeyCommand()
	mockIn, _, _ := tests.ApplyMockIO(deleteKeyCommand)

	yesF, _ := deleteKeyCommand.Flags().GetBool(flagYes)
	forceF, _ := deleteKeyCommand.Flags().GetBool(flagForce)
	require.False(t, yesF)
	require.False(t, forceF)

	fakeKeyName1 := "runDeleteCmd_Key1"
	fakeKeyName2 := "runDeleteCmd_Key2"
	kb, err := keyring.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), mockIn)
	require.NoError(t, err)
	t.Cleanup(func() {
		kb.Delete("runDeleteCmd_Key1", "", false) // nolint:errcheck
		kb.Delete("runDeleteCmd_Key2", "", false) // nolint:errcheck

	})
	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	t.Cleanup(cleanUp)
	viper.Set(flags.FlagHome, kbHome)

	// Now
	kb, err = keyring.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), kbHome, mockIn)
	require.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", "0", keyring.Secp256k1)
	require.NoError(t, err)

	_, err = kb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", "1", keyring.Secp256k1)
	require.NoError(t, err)

	err = runDeleteCmd(deleteKeyCommand, []string{"blah"})
	require.Error(t, err)
	require.Equal(t, "The specified item could not be found in the keyring", err.Error())

	// User confirmation missing
	err = runDeleteCmd(deleteKeyCommand, []string{fakeKeyName1})
	require.Error(t, err)
	require.Equal(t, "EOF", err.Error())

	{
		_, err = kb.Get(fakeKeyName1)
		require.NoError(t, err)

		// Now there is a confirmation
		viper.Set(flagYes, true)
		require.NoError(t, runDeleteCmd(deleteKeyCommand, []string{fakeKeyName1}))

		_, err = kb.Get(fakeKeyName1)
		require.Error(t, err) // Key1 is gone
	}

	viper.Set(flagYes, true)
	_, err = kb.Get(fakeKeyName2)
	require.NoError(t, err)
	err = runDeleteCmd(deleteKeyCommand, []string{fakeKeyName2})
	require.NoError(t, err)
	_, err = kb.Get(fakeKeyName2)
	require.Error(t, err) // Key2 is gone
}

func Test_confirmDeletion(t *testing.T) {
	type args struct {
		buf *bufio.Reader
	}

	answerYes := bufio.NewReader(strings.NewReader("y\n"))
	answerYes2 := bufio.NewReader(strings.NewReader("Y\n"))
	answerNo := bufio.NewReader(strings.NewReader("n\n"))
	answerInvalid := bufio.NewReader(strings.NewReader("245\n"))

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Y", args{answerYes}, false},
		{"y", args{answerYes2}, false},
		{"N", args{answerNo}, true},
		{"BAD", args{answerInvalid}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := confirmDeletion(tt.args.buf); (err != nil) != tt.wantErr {
				t.Errorf("confirmDeletion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
