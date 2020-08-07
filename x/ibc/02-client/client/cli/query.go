package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	types2 "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/gogo/protobuf/proto"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// GetCmdQueryClientStates defines the command to query all the light clients
// that this chain mantains.
func GetCmdQueryClientStates() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "states",
		Short:   "Query all available light clients",
		Long:    "Query all available light clients",
		Example: fmt.Sprintf("%s query %s %s states", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			clientStates, height, err := utils.QueryAllClientStates(clientCtx, page, limit)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)

			toPrint := make([]proto.Message, len(clientStates))
			for i, st := range clientStates {
				any, err := types2.NewAnyWithValue(st)
				if err != nil {
					return err
				}
				toPrint[i] = any
			}

			return clientCtx.PrintOutputArray(toPrint)
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of light clients to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of light clients to query for")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryClientState defines the command to query the state of a client with
// a given id as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryClientState() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "state [client-id]",
		Short:   "Query a client state",
		Long:    "Query stored client state",
		Example: fmt.Sprintf("%s query %s %s state [client-id]", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]
			if strings.TrimSpace(clientID) == "" {
				return errors.New("client ID can't be blank")
			}

			prove, _ := cmd.Flags().GetBool(flags.FlagProve)

			clientStateRes, err := utils.QueryClientState(clientCtx, clientID, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(clientStateRes.ProofHeight))
			return clientCtx.PrintOutputLegacy(clientStateRes)
		},
	}

	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryConsensusState defines the command to query the consensus state of
// the chain as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryConsensusState() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "consensus-state [client-id] [height]",
		Short:   "Query the consensus state of a client at a given height",
		Long:    "Query the consensus state for a particular light client at a given height",
		Example: fmt.Sprintf("%s query %s %s  consensus-state [client-id] [height]", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]
			if strings.TrimSpace(clientID) == "" {
				return errors.New("client ID can't be blank")
			}

			height, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("expected integer height, got: %s", args[1])
			}

			prove, _ := cmd.Flags().GetBool(flags.FlagProve)

			csRes, err := utils.QueryConsensusState(clientCtx, clientID, height, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(csRes.ProofHeight))
			return clientCtx.PrintOutputLegacy(csRes)
		},
	}

	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryHeader defines the command to query the latest header on the chain
func GetCmdQueryHeader() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "header",
		Short:   "Query the latest header of the running chain",
		Long:    "Query the latest Tendermint header of the running chain",
		Example: fmt.Sprintf("%s query %s %s  header", version.AppName, host.ModuleName, types.SubModuleName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			header, height, err := utils.QueryTendermintHeader(clientCtx)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutputLegacy(header)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdNodeConsensusState defines the command to query the latest consensus state of a node
// The result is feed to client creation
func GetCmdNodeConsensusState() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node-state",
		Short:   "Query a node consensus state",
		Long:    "Query a node consensus state. This result is feed to the client creation transaction.",
		Example: fmt.Sprintf("%s query %s %s node-state", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			state, height, err := utils.QueryNodeConsensusState(clientCtx)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutput(state)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
