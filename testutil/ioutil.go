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
// Returns an open file for the test to use.
func WriteToNewTempFile(t testing.TB, s string) *os.File {
	t.Helper()

	fp := TempFile(t)
	_, err := fp.WriteString(s)

	require.Nil(t, err)

	return fp
}

// TempFile returns a writable temporary file for the test to use.
func TempFile(t testing.TB) *os.File {
	t.Helper()

	fp, err := ioutil.TempFile(t.TempDir(), "")
	require.NoError(t, err)

	return fp
}
