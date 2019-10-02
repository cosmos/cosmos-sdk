package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	tmtypes "github.com/tendermint/tendermint/types"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
	"github.com/cosmos/cosmos-sdk/x/ibc/version"
)

// TODO: use Queriers

func mapping(cdc *codec.Codec, storeKey string, v int64) state.Mapping {
	prefix := version.Prefix(v)
	return state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
}

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
		Long: strings.TrimSpace(`Query stored client
		
$ <app>cli query ibc client state [id]
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			q := state.NewCLIQuerier(ctx)
			mapp := mapping(cdc, storeKey, version.Version)
			manager := types.NewManager(mapp)
			id := args[0]

			state, _, err := manager.State(id).ConsensusStateCLI(q)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, state))
			return nil
		},
	}
}

// GetCmdQueryRoot defines the command to query
func GetCmdQueryRoot(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "root",
		Short: "Query stored root",
		Long: strings.TrimSpace(`Query stored client
		
$ <app>cli query ibc client root [id] [height]
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			q := state.NewCLIQuerier(ctx)
			mapp := mapping(cdc, storeKey, version.Version)
			manager := types.NewManager(mapp)
			id := args[0]
			height, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			root, _, err := manager.State(id).RootCLI(q, height)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, root))
			return nil
		},
	}
}

// GetCmdQueryConsensusState defines the command to query the consensus state of
// the chain as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#query
func GetCmdQueryConsensusState(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "consensus-state",
		Short: "Query the latest consensus state of the running chain",
		Long: strings.TrimSpace(`Query consensus state
		
$ <app>cli query ibc client consensus-state
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			node, err := ctx.GetNode()
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

			state := tendermint.ConsensusState{
				ChainID:          commit.ChainID,
				Height:           uint64(commit.Height),
				Root:             merkle.NewRoot(commit.AppHash),
				NextValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, state))
			return nil
		},
	}
}

// GetCmdQueryPath defines the command to query the commitment path
func GetCmdQueryPath(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Query the commitment path of the running chain",
		Long: strings.TrimSpace(`Query the commitment path
		
$ <app>cli query ibc client path
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			mapp := mapping(cdc, storeName, version.Version)
			path := merkle.NewPrefix([][]byte{[]byte(storeName)}, mapp.PrefixBytes())
			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, path))
			return nil
		},
	}
}

// GetCmdQueryHeader defines the command to query the latest header on the chain
func GetCmdQueryHeader(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "header",
		Short: "Query the latest header of the running chain",
		Long: strings.TrimSpace(`Query the latest header
		
$ <app>cli query ibc client header
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			node, err := ctx.GetNode()
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

			nextValidators, err := node.Validators(&height)
			if err != nil {
				return err
			}

			header := tendermint.Header{
				SignedHeader:     commit.SignedHeader,
				ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
				NextValidatorSet: tmtypes.NewValidatorSet(nextValidators.Validators),
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, header))
			return nil
		},
	}
}
