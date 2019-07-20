package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
	"github.com/cosmos/cosmos-sdk/x/ibc/version"
)

func components(cdc *codec.Codec, storeKey string, version int64) (path merkle.Path, base state.Base) {
	prefix := []byte("v" + strconv.FormatInt(version, 10))
	path = merkle.NewPath([][]byte{[]byte(storeKey)}, prefix)
	base = state.NewBase(cdc, sdk.NewKVStoreKey(storeKey), prefix)
	return
}

func GetCmdQueryClient(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "client [clientid]",
		Short: "Query stored client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			path, base := components(cdc, storeKey, version.Version)
			man := client.NewManager(base)
			id := args[0]

			fmt.Println(string(man.CLIObject(path, id).ConsensusStateKey))
			state, _, err := man.CLIObject(path, id).ConsensusState(ctx)
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
				Root:             merkle.NewRoot(commit.AppHash),
				NextValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, state))

			return nil
		},
	}
}

func QueryHeader(ctx context.CLIContext) (res tendermint.Header, err error) {
	node, err := ctx.GetNode()
	if err != nil {
		return
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return
	}

	height := info.Response.LastBlockHeight
	prevheight := height - 1

	commit, err := node.Commit(&height)
	if err != nil {
		return
	}

	validators, err := node.Validators(&prevheight)
	if err != nil {
		return
	}

	nextvalidators, err := node.Validators(&height)
	if err != nil {
		return
	}

	return tendermint.Header{
		SignedHeader:     commit.SignedHeader,
		ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
		NextValidatorSet: tmtypes.NewValidatorSet(nextvalidators.Validators),
	}, nil

}

func GetCmdQueryHeader(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "header",
		Short: "Query the latest header of the running chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			header, err := QueryHeader(ctx)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, header))

			return nil
		},
	}
}
