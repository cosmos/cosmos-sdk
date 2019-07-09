package cli

import (
	"errors"
	"io/ioutil"
	//	"os"

	"github.com/spf13/cobra"

	//	"github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	//	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	//	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

const (
	FlagStatePath            = "state"
	FlagClientID             = "client-id"
	FlagConnectionID         = "connection-id"
	FlagChannelID            = "channel-id"
	FlagCounterpartyID       = "counterparty-id"
	FlagCounterpartyClientID = "counterparty-client-id"
	FlagSourceNode           = "source-node"
)

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcTxCmd := &cobra.Command{
		Use:                        "ibc",
		Short:                      "IBC transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcTxCmd.AddCommand(cli.PostCommands(
		GetCmdCreateClient(cdc),
		GetCmdUpdateClient(cdc),
	)...)

	return ibcTxCmd
}

func GetCmdCreateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-client",
		Short: "create new client with a consensus state",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc)

			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var state client.ConsensusState
			if err := cdc.UnmarshalJSON(contents, &state); err != nil {
				return err
			}

			msg := client.MsgCreateClient{
				ClientID:       args[0],
				ConsensusState: state,
				Signer:         cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func GetCmdUpdateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-client",
		Short: "update existing client with a header",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc)

			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var header client.Header
			if err := cdc.UnmarshalJSON(contents, &header); err != nil {
				return err
			}

			msg := client.MsgUpdateClient{
				ClientID: args[0],
				Header:   header,
				Signer:   cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// Copied from client/context/query.go
func query(ctx context.CLIContext, key []byte) ([]byte, merkle.Proof, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, merkle.Proof{}, err
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: ctx.Height,
		Prove:  true,
	}

	result, err := node.ABCIQueryWithOptions("/store/ibc/key", key, opts)
	if err != nil {
		return nil, merkle.Proof{}, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return nil, merkle.Proof{}, errors.New(resp.Log)
	}

	return resp.Value, merkle.Proof{
		Key:   key,
		Proof: resp.Proof,
	}, nil
}
