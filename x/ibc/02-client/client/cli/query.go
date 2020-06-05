package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// GetCmdQueryClientStates defines the command to query all the light clients
// that this chain mantains.
func GetCmdQueryClientStates(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "states",
		Short: "Query all available light clients",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all available light clients

Example:
$ %s query ibc client states
		`, version.ClientName),
		),
		Example: fmt.Sprintf("%s query ibc client states", version.ClientName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc)
			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			clientStates, height, err := utils.QueryAllClientStates(clientCtx, page, limit)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutput(clientStates)
		},
	}
	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of light clients to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of light clients to query for")
	return cmd
}

// GetCmdQueryClientState defines the command to query the state of a client with
// a given id as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryClientState(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state [client-id]",
		Short: "Query a client state",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query stored client state

Example:
$ %s query ibc client state [client-id]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc)
			clientID := args[0]
			if strings.TrimSpace(clientID) == "" {
				return errors.New("client ID can't be blank")
			}

			prove := viper.GetBool(flags.FlagProve)

			clientStateRes, err := utils.QueryClientState(clientCtx, clientID, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(clientStateRes.ProofHeight))
			return clientCtx.PrintOutput(clientStateRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	return cmd
}

// GetCmdQueryConsensusState defines the command to query the consensus state of
// the chain as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryConsensusState(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "consensus-state [client-id] [height]",
		Short:   "Query the consensus state of a client at a given height",
		Long:    "Query the consensus state for a particular light client at a given height",
		Example: fmt.Sprintf("%s query ibc client consensus-state [client-id] [height]", version.ClientName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc)
			clientID := args[0]
			if strings.TrimSpace(clientID) == "" {
				return errors.New("client ID can't be blank")
			}

			height, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("expected integer height, got: %s", args[1])
			}

			prove := viper.GetBool(flags.FlagProve)

			csRes, err := utils.QueryConsensusState(clientCtx, clientID, height, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(csRes.ProofHeight))
			return clientCtx.PrintOutput(csRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	return cmd
}

// GetCmdQueryHeader defines the command to query the latest header on the chain
func GetCmdQueryHeader(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "header",
		Short:   "Query the latest header of the running chain",
		Long:    "Query the latest Tendermint header of the running chain",
		Example: fmt.Sprintf("%s query ibc client header", version.ClientName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc)

			header, height, err := utils.QueryTendermintHeader(clientCtx)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutput(header)
		},
	}
}

// GetCmdNodeConsensusState defines the command to query the latest consensus state of a node
// The result is feed to client creation
func GetCmdNodeConsensusState(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "node-state",
		Short: "Query a node consensus state",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a node consensus state. This result is feed to the client creation transaction.

Example:
$ %s query ibc client node-state
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc)

			state, height, err := utils.QueryNodeConsensusState(clientCtx)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutput(state)
		},
	}
}

// GetCmdQueryPath defines the command to query the commitment path.
func GetCmdQueryPath(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Query the commitment path of the running chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			clienCtx := client.NewContext().WithCodec(cdc)
			path := commitmenttypes.NewMerklePrefix([]byte("ibc"))
			return clienCtx.PrintOutput(path)
		},
	}
}
