package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	query "github.com/cosmos/cosmos-sdk/x/metadata/querier"
	"github.com/cosmos/cosmos-sdk/x/metadata/types"
	"github.com/spf13/cobra"
)

func QueryTxCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the metadata module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetCmdShowMetadata(),
	)
	return queryCmd
}

func GetCmdShowMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Short:   "",
		Long:    "",
		Aliases: []string{"p"},
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := query.NewQueryClient(clientCtx)
			res, err := queryClient.Params(cmd.Context(), &query.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
