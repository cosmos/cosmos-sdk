package keys

import (
	"bufio"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
)

func Test_runDeleteCmd(t *testing.T) {
	deleteKeyCommand := deleteKeyCommand()

	yesF, _ := deleteKeyCommand.Flags().GetBool(flagYes)
	forceF, _ := deleteKeyCommand.Flags().GetBool(flagForce)
	assert.False(t, yesF)
	assert.False(t, forceF)

	fakeKeyName1 := "runDeleteCmd_Key1"
	fakeKeyName2 := "runDeleteCmd_Key2"

	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	defer cleanUp()
	viper.Set(flags.FlagHome, kbHome)

	// Now
	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", 0, 0)
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", 0, 1)
	assert.NoError(t, err)

	err = runDeleteCmd(deleteKeyCommand, []string{"blah"})
	require.Error(t, err)
	require.Equal(t, "Key blah not found", err.Error())

	// User confirmation missing
	err = runDeleteCmd(deleteKeyCommand, []string{fakeKeyName1})
	require.Error(t, err)
	require.Equal(t, "EOF", err.Error())

	{
		_, err = kb.Get(fakeKeyName1)
		require.NoError(t, err)

		// Now there is a confirmation
		mockIn, _, _ := tests.ApplyMockIO(deleteKeyCommand)
		mockIn.Reset("y\n")
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

	// TODO: Write another case for !keys.Local
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
		t.Run(tt.name, func(t *testing.T) {
			if err := confirmDeletion(tt.args.buf); (err != nil) != tt.wantErr {
				t.Errorf("confirmDeletion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
