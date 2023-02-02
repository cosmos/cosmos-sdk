package main

import (
	"bytes"
	"context"
	"testing"

	"cosmossdk.io/tools/cosmovisor"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand_Error(t *testing.T) {
	logger := cosmovisor.NewLogger()

	rootCmd.SetArgs([]string{"version"})

	out := bytes.NewBufferString("")
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	ctx := context.WithValue(context.Background(), cosmovisor.LoggerKey, logger)

	require.Error(t, rootCmd.ExecuteContext(ctx))
	require.Contains(t, out.String(), "DAEMON_NAME is not set")
}
