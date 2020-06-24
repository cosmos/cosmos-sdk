package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// GetCmdQueryChannel defines the command to query a channel end
func GetCmdQueryChannel(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "end [port-id] [channel-id]",
		Short: "Query a channel end",
		Long:  "Query an IBC channel end from a port and channel identifiers",
		Example: fmt.Sprintf(
			"%s query %s %s end [port-id] [channel-id]", version.ClientName, host.ModuleName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx = clientCtx.Init()

			portID := args[0]
			channelID := args[1]
			prove := viper.GetBool(flags.FlagProve)

			channelRes, err := utils.QueryChannel(clientCtx, portID, channelID, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(channelRes.ProofHeight))
			return clientCtx.PrintOutput(channelRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	return cmd
}

// GetCmdQueryChannelClientState defines the command to query a client state from a channel
func GetCmdQueryChannelClientState(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "client-state [port-id] [channel-id]",
		Short:   "Query the client state associated with a channel",
		Long:    "Query the client state associated with a channel, by providing its port and channel identifiers.",
		Example: fmt.Sprintf("%s query ibc channel client-state [port-id] [channel-id]", version.ClientName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx = clientCtx.Init()

			portID := args[0]
			channelID := args[1]

			clientStateRes, height, err := utils.QueryChannelClientState(clientCtx, portID, channelID)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutput(clientStateRes)
		},
	}
	return cmd
}

// GetCmdQueryNextSequence defines the command to query a next receive sequence for a given channel
func GetCmdQueryNextSequence(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next-sequence-receive [port-id] [channel-id]",
		Short: "Query a next receive sequence",
		Long: "Query the next receive sequence for a given channel"
		Example: fmt.Sprintf(
			"%s query %s %s next-sequence-receive [port-id] [channel-id]", version.ClientName, host.ModuleName, types.SubModuleName,
		),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx = clientCtx.Init()

			portID := args[0]
			channelID := args[1]
			prove := viper.GetBool(flags.FlagProve)

			sequenceRes, err := utils.QueryNextSequenceReceive(clientCtx, portID, channelID, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(sequenceRes.ProofHeight))
			return clientCtx.PrintOutput(sequenceRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	return cmd
}