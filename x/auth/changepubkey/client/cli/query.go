package cli

import (
	"context"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
)

// GetQueryCmd returns the transaction commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the auth module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetPubKeyHistoryCmd(),
		GetPubKeyHistoricalEntryCmd(),
		GetLastPubKeyHistoricalEntryCmd(),
		GetCurrentPubKeyEntryCmd(),
	)

	return cmd
}

// GetPubKeyHistoryCmd returns a query account pubkey history
func GetPubKeyHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubkey_history [address]",
		Short: "Query for pubkey_history by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PubKeyHistory(context.Background(), &types.QueryPubKeyHistoryRequest{Address: key.String()})
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetPubKeyHistoricalEntryCmd returns a query account pubkey history
func GetPubKeyHistoricalEntryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical_entry [address] [time]",
		Short: "Query for pubkey historical entry by address and time",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			timestamp, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			time := time.Unix(timestamp, 0)

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PubKeyHistoricalEntry(context.Background(), &types.QueryPubKeyHistoricalEntryRequest{
				Address: key.String(),
				Time:    time,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetLastPubKeyHistoricalEntryCmd returns a query account pubkey history
func GetLastPubKeyHistoricalEntryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last_pubkey_historical_entry [address]",
		Short: "Query for last pubkey historical entry by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.LastPubKeyHistoricalEntry(context.Background(), &types.QueryLastPubKeyHistoricalEntryRequest{
				Address: key.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCurrentPubKeyEntryCmd returns a query account pubkey history
func GetCurrentPubKeyEntryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current_pubkey_entry [address]",
		Short: "Query for current pubkey entry by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.CurrentPubKeyEntry(context.Background(), &types.QueryCurrentPubKeyEntryRequest{
				Address: key.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
