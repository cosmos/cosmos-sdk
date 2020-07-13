package testutil

import (
	"bytes"
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

// CallCliCmd calls theCmd cobra command and returns the output in bytes.
func CallCliCmd(clientCtx client.Context, theCmd func() *cobra.Command, extraArgs []string) ([]byte, error) {
	buf := new(bytes.Buffer)
	clientCtx = clientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd := theCmd()
	cmd.SetErr(buf)
	cmd.SetOut(buf)

	cmd.SetArgs(extraArgs)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}
