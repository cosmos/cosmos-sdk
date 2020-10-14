package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/convo/types"
)

// GetCmdQueryPendingMessage defines the command to query a pending (receipt unconfirmed) message from a sender to a receiver
// over a given channel
func GetCmdQueryPendingMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pending-message [sender] [receiver] [channel]",
		Short:   "Query a sender's pending message to a receiver over a given channel",
		Long:    "Query a sender's pending message to a receiver over a given channel",
		Example: fmt.Sprintf("%s query ibc-convo pending [sender] [receiver] [channel]", version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryPendingMessageRequest{
				Sender:   args[0],
				Receiver: args[1],
				Channel:  args[2],
			}

			res, err := queryClient.PendingMessage(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryInboxMessage defines the command to query an inbox message sent to a receiver by the sender
// over a given channel
func GetCmdQueryInboxMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inbox-message [receiver] [sender] [channel]",
		Short:   "Query a receiver's inbox message from a sender over a given channel",
		Long:    "Query a receiver's inbox message from a sender over a given channel",
		Example: fmt.Sprintf("%s query ibc-convo inbox [receiver] [sender] [channel]", version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryInboxMessageRequest{
				Receiver: args[0],
				Sender:   args[1],
				Channel:  args[2],
			}

			res, err := queryClient.InboxMessage(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryOutboxMessage defines the command to query an outbox message sent by a sender to the receiver
// over a given channel
func GetCmdQueryOutboxMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "outbox-message [sender] [receiver] [channel]",
		Short:   "Query a receiver's outbox message from a sender over a given channel",
		Long:    "Query a receiver's outbox message from a sender over a given channel",
		Example: fmt.Sprintf("%s query ibc-convo outbox [receiver] [sender] [channel]", version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOutboxMessageRequest{
				Receiver: args[0],
				Sender:   args[1],
				Channel:  args[2],
			}

			res, err := queryClient.OutboxMessage(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
