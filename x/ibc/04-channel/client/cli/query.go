package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/utils"
)

// GetCmdQueryChannel defines the command to query a channel end
func GetCmdQueryChannel(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "end [port-id] [channel-id]",
		Short: "Query a channel end",
		Long: strings.TrimSpace(fmt.Sprintf(`Query an IBC channel end
		
Example:
$ %s query ibc channel end [port-id] [channel-id]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			ch, err := utils.QueryChannel(cliCtx, args[0], args[1], queryRoute, true)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(ch)
		},
	}
	return cmd
}
