package testutil

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// NewTestCaseDir creates a new temporary directory for a test case.
// Returns the directory path and a cleanup function.
// nolint: errcheck
func NewTestCaseDir(t testing.TB) (string, func()) {
	dir, err := ioutil.TempDir("", t.Name()+"_")
	require.NoError(t, err)
	return dir, func() { os.RemoveAll(dir) }
}

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

// Write the given string to a new temporary file.
// Returns an open file and a clean up function that
// the caller must call to remove the file when it is
// no longer needed.
func WriteToNewTempFile(t testing.TB, s string) (*os.File, func()) {
	fp, err := ioutil.TempFile("", strings.ReplaceAll(t.Name(), "/", "_")+"_")
	require.Nil(t, err)

	_, err = fp.WriteString(s)
	require.Nil(t, err)

	return fp, func() { os.Remove(fp.Name()) }
}
