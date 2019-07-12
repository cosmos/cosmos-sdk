package tests

import (
	"bytes"
	"strings"

	"github.com/spf13/cobra"
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

// DONTCOVER
