package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

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
		Args:    cobra.NoArgs,
		Short:   "Query metadata params",
		Aliases: []string{"p"},
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				// return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			_, err = queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			// return clientCtx.PrintProto(&res.Params)
			return clientCtx.PrintProto(&types.Params{})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
