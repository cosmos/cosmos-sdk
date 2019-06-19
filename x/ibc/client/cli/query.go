package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/keeper"
)

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcQueryCmd := &cobra.Command{
		Use:                        "ibc",
		Short:                      "IBC query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryConsensusState(cdc),
		GetCmdQueryHeader(cdc),
		GetCmdQueryClient(cdc),
		GetCmdQueryConnection(cdc),
		GetCmdQueryChannel(cdc),
	)...)
	return ibcQueryCmd
}

func GetCmdQueryClient(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "client",
		Short: "Query stored client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			keeper := ibc.DummyKeeper()

			var state ibc.ConsensusState
			statebz, _, err := query(ctx, keeper.Client.Object(args[0]).Key())
			if err != nil {
				return err
			}
			cdc.MustUnmarshalBinaryBare(statebz, &state)

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, state))

			return nil
		},
	}
}

func GetCmdQueryConsensusState(cdc *codec.Codec) *cobra.Command {
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
				Root:             []byte(commit.AppHash),
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

func GetCmdQueryConnection(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "connection",
		Short: "Query an existing connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			keeper := ibc.DummyKeeper()

			var conn ibc.Connection
			connbz, _, err := query(ctx, keeper.Connection.Object(args[0]).Key())
			if err != nil {
				return err
			}
			cdc.MustUnmarshalBinaryBare(connbz, &conn)

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, conn))

			return nil
		},
	}
}

func GetCmdQueryChannel(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "channel",
		Short: "Query an existing channel",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			keeper := ibc.DummyKeeper()

			var conn ibc.Channel
			connbz, _, err := query(ctx, keeper.Channel.Object(args[0], args[1]).Key())
			if err != nil {
				return err
			}
			cdc.MustUnmarshalBinaryBare(connbz, &conn)

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, conn))

			return nil
		},
	}
}

func GetCmdQuerySendSequence(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "send-sequence",
		Short: "Query the send sequence of a channel",
		Args:  cobra.ExactArgs(),
		RunE: func(cmd *cobra.Command, args []string) error {

		},
	}
}

func GetCmdQueryReceiveSequence(cdc *codec.Codec) *cobra.Command {

}

func GetCmdQueryPacket(cdc *codec.Codec) *cobra.Command {
}
