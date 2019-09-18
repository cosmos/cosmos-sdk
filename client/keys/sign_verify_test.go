package keys

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
)

func Test_runSignCmd(t *testing.T) {
	signCmd := signCommand()
	// err := runSignCmd(signCmd, []string{"invalid", "invalid"})
	// require.Contains(t, err.Error(), "no such file or directory")

	// Prepare a plain text doc
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	ioutil.WriteFile(tmpfile.Name(), []byte(`this is
an example`), 0644)
	require.NoError(t, err)
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Prepare a key base
	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	defer cleanUp()
	viper.Set(flags.FlagHome, kbHome)
	viper.Set(cli.OutputFlag, OutputFormatText)

	// Initialise keybase
	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	_, err = kb.CreateAccount("key1", tests.TestMnemonic, "", "test1234", 0, 0)
	assert.NoError(t, err)

	// Mock standard streams
	mockIn, mockOut, _ := tests.ApplyMockIO(signCmd)
	mockIn.Reset("test1234\n")
	require.NoError(t, runSignCmd(signCmd, []string{"key1", tmpfile.Name()}))

	signedDocBytes := mockOut.Bytes()
	var signedDoc signedText
	require.NoError(t, UnmarshalJSON(signedDocBytes, &signedDoc))

	// Prepare a signed doc file
	outTmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	ioutil.WriteFile(outTmpFile.Name(), signedDocBytes, 0644)
	require.NoError(t, err)
	outTmpFile.Close()
	defer os.Remove(outTmpFile.Name())

	// Verify
	verifyCommand := verifyCommand()
	require.NoError(t, runVerifyCmd(verifyCommand, []string{outTmpFile.Name()}))

	// Prepare a file with corrupted signature
	signedDoc.Text = "this is an example"
	signedDocBytes, err = MarshalJSON(signedDoc)
	require.NoError(t, err)

	// Verification fails
	outTmpFile, err = ioutil.TempFile("", "")
	require.NoError(t, err)
	ioutil.WriteFile(outTmpFile.Name(), signedDocBytes, 0644)
	require.NoError(t, err)
	outTmpFile.Close()
	defer os.Remove(outTmpFile.Name())

	require.Error(t, runVerifyCmd(verifyCommand, []string{outTmpFile.Name()}))
}
