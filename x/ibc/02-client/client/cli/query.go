package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// GetQueryCmd returns the query commands for IBC clients
func GetQueryCmd(queryRouter string, cdc *codec.Codec) *cobra.Command {
	ics02ClientQueryCmd := &cobra.Command{
		Use:                        "client",
		Short:                      "IBC client query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ics02ClientQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryConsensusState(queryRouter, cdc),
		GetCmdQueryPath(queryRouter, cdc),
		GetCmdQueryHeader(cdc),
		GetCmdQueryClientState(queryRouter, cdc),
		GetCmdQueryRoot(queryRouter, cdc),
	)...)
	return ics02ClientQueryCmd
}

// GetCmdQueryClientState defines the command to query the state of a client with
// a given id as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryClientState(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
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
			id := args[0]

			bz, err := cdc.MarshalJSON(types.NewQueryClientStateParams(id))
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(types.ClientStatePath(id), bz)
			if err != nil {
				return err
			}

			var clientState types.ClientState
			if err := cdc.UnmarshalJSON(res, &clientState); err != nil {
				return err
			}

			return cliCtx.PrintOutput(clientState)
		},
	}
}

// GetCmdQueryRoot defines the command to query
func GetCmdQueryRoot(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "root [client-id] [height]",
		Short: "Query stored root",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a stored commitment root at a specific height for a particular client

Example:
$ %s query ibc client root [client-id] [height]
`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			id := args[0]
			height, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.NewQueryCommitmentRootParams(id, height))
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(types.RootPath(id, height), bz)
			if err != nil {
				return err
			}

			var root ics23.Root
			if err := cdc.UnmarshalJSON(res, &root); err != nil {
				return err
			}

			return cliCtx.PrintOutput(root)
		},
	}
}

// GetCmdQueryConsensusState defines the command to query the consensus state of
// the chain as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryConsensusState(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
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
			id := args[0]

			bz, err := cdc.MarshalJSON(types.NewQueryClientStateParams(id))
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(types.ConsensusStatePath(id), bz)
			if err != nil {
				return err
			}

			var consensusState exported.ConsensusState
			if err := cdc.UnmarshalJSON(res, &consensusState); err != nil {
				return err
			}

			return cliCtx.PrintOutput(consensusState)
		},
	}
}

// GetCmdQueryPath defines the command to query the commitment path
func GetCmdQueryPath(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Query the commitment path of the running chain",
		Long: strings.TrimSpace(fmt.Sprintf(`Query the commitment path
		
Example:
$ %s query ibc client path
		`, version.ClientName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// TODO: get right path
			res, _, err := cliCtx.Query("")
			if err != nil {
				return err
			}

			var path merkle.Prefix
			if err := cdc.UnmarshalJSON(res, &path); err != nil {
				return err
			}

			return cliCtx.PrintOutput(path)
		},
	}
}

// GetCmdQueryHeader defines the command to query the latest header on the chain
// TODO: do we really need this cmd ??
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

			nextValidators, err := node.Validators(&height)
			if err != nil {
				return err
			}

			header := tendermint.Header{
				SignedHeader:     commit.SignedHeader,
				ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
				NextValidatorSet: tmtypes.NewValidatorSet(nextValidators.Validators),
			}

			return cliCtx.PrintOutput(header)
		},
	}
}
