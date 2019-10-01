package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/mock/types"
)

func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "ibcmockrecv",
		Short: "Querying commands for the ibcmockrecv module",
		RunE:  client.ValidateCmd,
	}

	queryCmd.AddCommand(client.GetCommands(
		GetCmdQuerySequence(queryRoute, cdc),
	)...)

	return queryCmd
}

func GetCmdQuerySequence(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "sequence [channel-id]",
		Short: "Query the current sequence for the channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			val, _, err := ctx.QueryStore(types.SequenceKey(args[0]), storeName)
			if err != nil {
				return err
			}

			var res uint64
			if val == nil {
				res = 0
			} else {
				cdc.MustUnmarshalBinaryBare(val, &res)
			}
			fmt.Println(res)

			return nil
		},
	}
}