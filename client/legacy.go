package client

import "github.com/spf13/cobra"

// WithLegacyAminoJSON configures the command's client Context to use the legacy
// amino *codec.Codec as its JSONMarshaler. This is to be used for commands that
// need to always use amino JSON for legacy compatibility reasons.
func WithLegacyAminoJSON(cmd *cobra.Command) {
	AddPreRunEHook(cmd, func(cmd *cobra.Command, args []string) error {
		clientCtx := GetClientContextFromCmd(cmd)
		clientCtx = clientCtx.WithJSONMarshaler(clientCtx.Codec)
		return SetCmdClientContext(cmd, clientCtx)
	})
}
