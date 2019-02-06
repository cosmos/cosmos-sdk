package keys

import (
	"bufio"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/stretchr/testify/assert"
)

func Test_runDeleteCmd(t *testing.T) {
	deleteKeyCommand := deleteKeyCommand()

	yesF, _ := deleteKeyCommand.Flags().GetBool(flagYes)
	forceF, _ := deleteKeyCommand.Flags().GetBool(flagForce)
	assert.False(t, yesF)
	assert.False(t, forceF)

	fakeKeyName := "runDeleteCmd1"

	kb := keys.New(dbm.NewMemDB())
	_, err := kb.CreateAccount(fakeKeyName,
		tests.TestMnemonic, "", "", 0, 0)
	assert.NoError(t, err)

	err = runDeleteCmd(deleteKeyCommand, []string{})
	require.Error(t, err)
	require.Equal(t, "a name must be provided", err.Error())

	err = runDeleteCmd(deleteKeyCommand, []string{"blah"})
	require.Error(t, err)
	require.Equal(t, "file missing [file=MANIFEST-000000]", err.Error())

	// Now add a keybase
	SetKeyBase(kb)
	err = runDeleteCmd(deleteKeyCommand, []string{"blah"})
	require.Error(t, err)
	require.Equal(t, "Key blah not found", err.Error())

	// User confirmation
	// TODO: Mock stdin?
	err = runDeleteCmd(deleteKeyCommand, []string{fakeKeyName})
	require.Error(t, err)
	require.Equal(t, "EOF", err.Error())

	viper.Set(flagYes, true)
	_, err = kb.Get(fakeKeyName)
	require.NoError(t, err)
	err = runDeleteCmd(deleteKeyCommand, []string{fakeKeyName})
	require.NoError(t, err)
	_, err = kb.Get(fakeKeyName)
	require.Error(t, err)

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
