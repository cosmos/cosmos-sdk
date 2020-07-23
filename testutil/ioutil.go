package testutil

import (
	"bytes"
	"io"
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
	dir, err := ioutil.TempDir("", strings.ReplaceAll(t.Name(), "/", "_")+"_")
	require.NoError(t, err)
	return dir, func() { os.RemoveAll(dir) }
}

// BufferReader is implemented by types that read from a string buffer.
type BufferReader interface {
	io.Reader
	Reset(string)
}

// BufferWriter is implemented by types that write to a buffer.
type BufferWriter interface {
	io.Writer
	Reset()
	Bytes() []byte
	String() string
}

// ApplyMockIO replaces stdin/out/err with buffers that can be used during testing.
// Returns an input BufferReader and an output BufferWriter.
func ApplyMockIO(c *cobra.Command) (BufferReader, BufferWriter) {
	mockIn := strings.NewReader("")
	mockOut := bytes.NewBufferString("")

	c.SetIn(mockIn)
	c.SetOut(mockOut)
	c.SetErr(mockOut)

	return mockIn, mockOut
}

// ApplyMockIODiscardOutputs replaces a cobra.Command output and error streams with a dummy io.Writer.
// Replaces and returns the io.Reader associated to the cobra.Command input stream.
func ApplyMockIODiscardOutErr(c *cobra.Command) BufferReader {
	mockIn := strings.NewReader("")

	c.SetIn(mockIn)
	c.SetOut(ioutil.Discard)
	c.SetErr(ioutil.Discard)

	return mockIn
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
