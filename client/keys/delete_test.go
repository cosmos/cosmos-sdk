package keys

import (
	"bufio"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runDeleteCmd(t *testing.T) {
	runningUnattended := isRunningUnattended()
	deleteKeyCommand := DeleteKeyCommand()
	mockIn, _, _ := tests.ApplyMockIO(deleteKeyCommand)

	yesF, _ := deleteKeyCommand.Flags().GetBool(flagYes)
	forceF, _ := deleteKeyCommand.Flags().GetBool(flagForce)
	require.False(t, yesF)
	require.False(t, forceF)

	fakeKeyName1 := "runDeleteCmd_Key1"
	fakeKeyName2 := "runDeleteCmd_Key2"
	if !runningUnattended {
		kb, err := keys.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), mockIn)
		require.NoError(t, err)
		defer func() {
			kb.Delete("runDeleteCmd_Key1", "", false)
			kb.Delete("runDeleteCmd_Key2", "", false)

		}()
	}
	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	defer cleanUp()
	viper.Set(flags.FlagHome, kbHome)

	// Now
	kb, err := keys.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), kbHome, mockIn)
	require.NoError(t, err)
	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	_, err = kb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", sdk.FullFundraiserPath, keys.Secp256k1)
	require.NoError(t, err)

	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	_, err = kb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", "m/44'/118'/0'/0/1", keys.Secp256k1)
	require.NoError(t, err)

	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	err = runDeleteCmd(deleteKeyCommand, []string{"blah"})
	require.Error(t, err)
	require.Equal(t, "The specified item could not be found in the keyring", err.Error())

	// User confirmation missing
	err = runDeleteCmd(deleteKeyCommand, []string{fakeKeyName1})
	require.Error(t, err)
	if runningUnattended {
		require.Equal(t, "aborted", err.Error())
	} else {
		require.Equal(t, "EOF", err.Error())
	}

	{
		if runningUnattended {
			mockIn.Reset("testpass1\n")
		}
		_, err = kb.Get(fakeKeyName1)
		require.NoError(t, err)

		// Now there is a confirmation
		viper.Set(flagYes, true)
		if runningUnattended {
			mockIn.Reset("testpass1\ntestpass1\n")
		}
		require.NoError(t, runDeleteCmd(deleteKeyCommand, []string{fakeKeyName1}))

		_, err = kb.Get(fakeKeyName1)
		require.Error(t, err) // Key1 is gone
	}

	viper.Set(flagYes, true)
	if runningUnattended {
		mockIn.Reset("testpass1\n")
	}
	_, err = kb.Get(fakeKeyName2)
	require.NoError(t, err)
	if runningUnattended {
		mockIn.Reset("testpass1\ny\ntestpass1\n")
	}
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
