package tests

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// ApplyMockIO replaces stdin/out/err with buffers that can be used during testing.
func ApplyMockIO(c *cobra.Command) (*strings.Reader, *bytes.Buffer, *bytes.Buffer) {
	mockIn := strings.NewReader("")
	mockOut := bytes.NewBufferString("")
	mockErr := bytes.NewBufferString("")
	c.SetIn(mockIn)
	c.SetOut(mockOut)
	c.SetErr(mockErr)
	return mockIn, mockOut, mockErr
}

// Write the given string to a new temporary file
func WriteToNewTempFile(t require.TestingT, s string) (*os.File, func()) {
	fp, err := ioutil.TempFile(os.TempDir(), "cosmos_cli_test_")
	require.Nil(t, err)

	_, err = fp.WriteString(s)
	require.Nil(t, err)

	return fp, func() { os.Remove(fp.Name()) }
}

// DONTCOVER
