package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// GetQueryCmd returns the query commands for IBC clients
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	ics02ClientQueryCmd := &cobra.Command{
		Use:                        "client",
		Short:                      "IBC client query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ics02ClientQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryConsensusState(queryRoute, cdc),
		GetCmdQueryHeader(cdc),
		GetCmdQueryClientState(queryRoute, cdc),
		GetCmdQueryRoot(queryRoute, cdc),
		GetCmdNodeConsensusState(queryRoute, cdc),
		GetCmdQueryPath(queryRoute, cdc),
	)...)
	return ics02ClientQueryCmd
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
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			clientID := args[0]
			if strings.TrimSpace(clientID) == "" {
				return errors.New("client ID can't be blank")
			}

			bz, err := cdc.MarshalJSON(types.NewQueryClientStateParams(clientID))
			if err != nil {
				return err
			}

			req := abci.RequestQuery{
				Path:  fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryClientState),
				Data:  bz,
				Prove: viper.GetBool(flags.FlagProve),
			}

			res, err := cliCtx.QueryABCI(req)
			if err != nil {
				return err
			}

			var clientState types.State
			if err := cdc.UnmarshalJSON(res.Value, &clientState); err != nil {
				return err
			}

			clientStateRes := types.NewClientStateResponse(clientID, clientState, res.Proof, res.Height)
			return cliCtx.PrintOutput(clientStateRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	return cmd
}

// GetCmdQueryRoot defines the command to query
func GetCmdQueryRoot(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "root [client-id] [height]",
		Short: "Query a verified commitment root",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query an already verified commitment root at a specific height for a particular client

Example:
$ %s query ibc client root [client-id] [height]
`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			clientID := args[0]
			if strings.TrimSpace(clientID) == "" {
				return errors.New("client ID can't be blank")
			}

			height, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("expected integer height, got: %v", args[1])
			}

			bz, err := cdc.MarshalJSON(types.NewQueryCommitmentRootParams(clientID, height))
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryVerifiedRoot), bz)
			if err != nil {
				return err
			}

			var root commitment.RootI
			if err := cdc.UnmarshalJSON(res, &root); err != nil {
				return err
			}

			return cliCtx.PrintOutput(root)
		},
	}
}

// GetCmdQueryConsensusState defines the command to query the consensus state of
// the chain as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryConsensusState(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consensus-state [client-id]",
		Short: "Query the latest consensus state of the client",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the consensus state for a particular client

Example:
$ %s query ibc client consensus-state [client-id]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			clientID := args[0]
			if strings.TrimSpace(clientID) == "" {
				return errors.New("client ID can't be blank")
			}

			csRes, err := utils.QueryConsensusStateProof(cliCtx, cdc, queryRoute, clientID)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(csRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	return cmd
}

// GetCmdQueryHeader defines the command to query the latest header on the chain
func GetCmdQueryHeader(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "header",
		Short: "Query the latest header of the running chain",
		Long: strings.TrimSpace(fmt.Sprintf(`Query the latest Tendermint header
		
Example:
$ %s query ibc client header
		`, version.ClientName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			header, err := utils.GetTendermintHeader(cliCtx)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(header)
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
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			node, err := cliCtx.GetNode()
			if err != nil {
				return err
			}

			info, err := node.ABCIInfo()
			if err != nil {
				return err
			}

			height := info.Response.LastBlockHeight
			prevHeight := height - 1

			commit, err := node.Commit(&height)
			if err != nil {
				return err
			}

			validators, err := node.Validators(&prevHeight)
			if err != nil {
				return err
			}

			state := tendermint.ConsensusState{
				ChainID:          commit.ChainID,
				Height:           uint64(commit.Height),
				Root:             commitment.NewRoot(commit.AppHash),
				NextValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
			}

			return cliCtx.PrintOutput(state)
		},
	}
}

// GetCmdQueryPath defines the command to query the commitment path.
func GetCmdQueryPath(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Query the commitment path of the running chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			path := commitment.NewPrefix([]byte("ibc"))
			return ctx.PrintOutput(path)
		},
	}
}
