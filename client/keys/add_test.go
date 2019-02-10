package keys

import (
	"bufio"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/tests"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/stretchr/testify/assert"
)

func Test_runAddCmdBasic(t *testing.T) {
	// Prepare a keybase
	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()

	// Missing input (enter password)
	assert.EqualError(t, runAddCmd(kbHome, []string{"keyname"}), "EOF")

	/// Test Text
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	cleanUp1 := client.OverrideStdin(bufio.NewReader(strings.NewReader("test1234\ntest1234\n")))
	defer cleanUp1()
	assert.NoError(t, runAddCmd(kbHome, []string{"keyname1"}))

	/// Test Text - Replace? >> FAIL
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	cleanUp2 := client.OverrideStdin(bufio.NewReader(strings.NewReader("test1234\ntest1234\n")))
	defer cleanUp2()
	assert.Error(t, runAddCmd(kbHome, []string{"keyname1"}))

	/// Test Text - Replace? Answer >> PASS
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	cleanUp3 := client.OverrideStdin(bufio.NewReader(strings.NewReader("y\ntest1234\ntest1234\n")))
	defer cleanUp3()
	assert.NoError(t, runAddCmd(kbHome, []string{"keyname1"}))

	// Check JSON
	viper.Set(cli.OutputFlag, OutputFormatJSON)
	// Now enter password
	cleanUp4 := client.OverrideStdin(bufio.NewReader(strings.NewReader("test1234\ntest1234\n")))
	defer cleanUp4()
	assert.NoError(t, runAddCmd(kbHome, []string{"keyname2"}))
}

type MockResponseWriter struct {
	dataHeaderStatus int
	dataBody         []byte
}

func (MockResponseWriter) Header() http.Header {
	panic("Unexpected call!")
}

func (w *MockResponseWriter) Write(data []byte) (int, error) {
	w.dataBody = append(w.dataBody, data...)
	return len(data), nil
}

func (w *MockResponseWriter) WriteHeader(statusCode int) {
	w.dataHeaderStatus = statusCode
}

func TestCheckAndWriteErrorResponse(t *testing.T) {
	mockRW := MockResponseWriter{}

	mockRW.WriteHeader(100)
	assert.Equal(t, 100, mockRW.dataHeaderStatus)

	detected := CheckAndWriteErrorResponse(&mockRW, http.StatusBadRequest, errors.New("some ERROR"))
	require.True(t, detected)
	require.Equal(t, http.StatusBadRequest, mockRW.dataHeaderStatus)
	require.Equal(t, "some ERROR", string(mockRW.dataBody[:]))

	mockRW = MockResponseWriter{}
	detected = CheckAndWriteErrorResponse(&mockRW, http.StatusBadRequest, nil)
	require.False(t, detected)
	require.Equal(t, 0, mockRW.dataHeaderStatus)
	require.Equal(t, "", string(mockRW.dataBody[:]))
}
