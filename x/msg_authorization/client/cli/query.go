package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	authorizationQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the msg authorization module",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query any 'Capability' (or 'nil'), with the expiration time, granted to the grantee by the granter for the provided msg type.
$ %[1]s query msg_authorization comosaddr98765678ijhb cosmosaddr876hjjhy88i hello`,
				version.AppName,
			),
		),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	authorizationQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryAuthorization(queryRoute, cdc),
	)...)

	return authorizationQueryCmd
}

// GetCmdQueryAuthorization implements the query authorizations command.
func GetCmdQueryAuthorization(storeName string, cdc *codec.Codec) *cobra.Command {
	//TODO update description
	cmd := &cobra.Command{
		Use:   "authorization [granter-address] [grantee-address]",
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

			queryClient.Authorization(
				context.Background(),
				&types.QueryAuthorizationRequest{
					GranterAddress: granterAddr,
					GranteeAddress: granteeAddr, MsgType: msgAuthorized,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(&res.Proposal)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
