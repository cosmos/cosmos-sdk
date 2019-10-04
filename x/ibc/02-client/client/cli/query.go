package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetQueryCmd returns the query commands for IBC clients
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcQueryCmd := &cobra.Command{
		Use:                        "client",
		Short:                      "IBC client query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcQueryCmd.AddCommand(cli.GetCommands(
		GetCmdQueryConsensusState(storeKey, cdc),
		GetCmdQueryPath(storeKey, cdc),
		GetCmdQueryHeader(cdc),
		GetCmdQueryClientState(storeKey, cdc),
		GetCmdQueryRoot(storeKey, cdc),
	)...)
	return ibcQueryCmd
}

// GetCmdQueryClientState defines the command to query the state of a client with
// a given id as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryClientState(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "state",
		Short: "Query a client state",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query stored client

Example:
$ %s query ibc client state [id]
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

			req := abci.RequestQuery{
				Path:  "/store/" + storeKey + "/key",
				Data:  bz,
				Prove: true,
			}

			return cliCtx.PrintOutput()
		},
	}
}

// GetCmdQueryRoot defines the command to query
func GetCmdQueryRoot(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "root",
		Short: "Query stored root",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query stored client

Example:
$ %s query ibc client root [id] [height]
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

			return cliCtx.PrintOutput()
		},
	}
}

// GetCmdQueryConsensusState defines the command to query the consensus state of
// the chain as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryConsensusState(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "consensus-state",
		Short: "Query the latest consensus state of the running chain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query consensus state

Example:
$ %s query ibc client consensus-state
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
			prevheight := height - 1
			commit, err := node.Commit(&height)
			if err != nil {
				return err
			}

			validators, err := node.Validators(&prevheight)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput()
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

			return cliCtx.PrintOutput()
		},
	}
}

// GetCmdQueryHeader defines the command to query the latest header on the chain
func GetCmdQueryHeader(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "header",
		Short: "Query the latest header of the running chain",
		Long: strings.TrimSpace(fmt.Sprintf(`Query the latest header
		
Example:
$ %s query ibc client header
		`, version.ClientName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			return cliCtx.PrintOutput()
		},
	}
}
