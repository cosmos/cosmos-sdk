package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	tmtypes "github.com/tendermint/tendermint/types"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func defaultRoot(storeKey string, root []byte) merkle.Root {
	return merkle.NewRoot(root, [][]byte{[]byte(storeKey)}, []byte("protocol"))
}

func defaultBase(cdc *codec.Codec) (state.Base, state.Base) {
	protocol := state.NewBase(cdc, nil, []byte("protocol"))
	free := state.NewBase(cdc, nil, []byte("free"))
	return protocol, free
}

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcQueryCmd := &cobra.Command{
		Use:                        "ibc",
		Short:                      "IBC query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcQueryCmd.AddCommand(cli.GetCommands(
		GetCmdQueryConsensusState(storeKey, cdc),
		GetCmdQueryHeader(cdc),
		GetCmdQueryClient(storeKey, cdc),
	)...)
	return ibcQueryCmd
}

func GetCmdQueryClient(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "client",
		Short: "Query stored client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			man := client.NewManager(defaultBase(cdc))
			root := defaultRoot(storeKey, nil)
			id := args[0]

			state, _, err := man.CLIObject(root, id).ConsensusState(ctx)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, state))

			return nil
		},
	}
}

func GetCmdQueryConsensusState(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "consensus-state",
		Short: "Query the latest consensus state of the running chain",
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
				Root:             defaultRoot(storeKey, []byte(commit.AppHash)),
				NextValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, state))

			return nil
		},
	}
}

func GetCmdQueryHeader(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "header",
		Short: "Query the latest header of the running chain",
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

			nextvalidators, err := node.Validators(&height)
			if err != nil {
				return err
			}

			header := tendermint.Header{
				SignedHeader:     commit.SignedHeader,
				ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
				NextValidatorSet: tmtypes.NewValidatorSet(nextvalidators.Validators),
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, header))

			return nil
		},
	}
}
