package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	authorizationQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the msg authorization module",
		Long:                       "",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	authorizationQueryCmd.AddCommand(
		GetCmdQueryAuthorization(),
	)

	return authorizationQueryCmd
}

// GetCmdQueryAuthorization implements the query authorizations command.
func GetCmdQueryAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorization [granter-addr] [grantee-addr] [msg-type]",
		Args:  cobra.ExactArgs(3),
		Short: "query authorization for a granter-grantee pair",
		Long:  "query authorization for a granter-grantee pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			granterAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			granteeAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msgAuthorized := args[2]

			res, err := queryClient.Authorization(
				context.Background(),
				&types.QueryAuthorizationRequest{
					GranterAddr: granterAddr.String(),
					GranteeAddr: granteeAddr.String(),
					MsgType:     msgAuthorized,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.Authorization)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
