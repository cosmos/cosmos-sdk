package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewQueryCmd returns the custom query commands of the auth module.
func NewQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Auth query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		NewRekeyedAccountsCmd(),
		NewPubKeyHistoryCmd(),
	)

	return queryCmd
}

// NewRekeyedAccountsCmd queries accounts controlled by a public key's natural
// address.
func NewRekeyedAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rekeyed-accounts [address]",
		Short: "Query the accounts whose current public key has the given natural address",
		Long:  "Query the accounts whose current public key has the given natural address. Use it to find the account a key controls after a pubkey change.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageFlags, err := client.FlagSetWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}
			pageReq, err := client.ReadPageRequest(pageFlags)
			if err != nil {
				return err
			}

			res, err := types.NewQueryClient(clientCtx).RekeyedAccounts(cmd.Context(), &types.QueryRekeyedAccountsRequest{
				Address:    args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "rekeyed accounts")

	return cmd
}

// NewPubKeyHistoryCmd queries the public key rotation history of an account.
func NewPubKeyHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubkey-history [address]",
		Short: "Query the public key rotation history of an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageFlags, err := client.FlagSetWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}
			pageReq, err := client.ReadPageRequest(pageFlags)
			if err != nil {
				return err
			}

			res, err := types.NewQueryClient(clientCtx).PubKeyHistory(cmd.Context(), &types.QueryPubKeyHistoryRequest{
				Address:    args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "public key history")

	return cmd
}
