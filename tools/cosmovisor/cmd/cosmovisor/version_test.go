package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
)

func TestVersionCommand_Error(t *testing.T) {
	logger := log.NewTestLogger(t).With(log.ModuleKey, "cosmovisor")

	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"version"})

	out := bytes.NewBufferString("")
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	ctx := context.WithValue(context.Background(), log.ContextKey, logger)

	require.Error(t, rootCmd.ExecuteContext(ctx))
	require.Contains(t, out.String(), "DAEMON_NAME is not set")
}
