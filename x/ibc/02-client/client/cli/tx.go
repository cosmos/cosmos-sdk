package cli

import (
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// ICS02 Client CLI flags
const (
	FlagStatePath            = "state"
	FlagClientID             = "client-id"
	FlagConnectionID         = "connection-id"
	FlagChannelID            = "channel-id"
	FlagCounterpartyID       = "counterparty-id"
	FlagCounterpartyClientID = "counterparty-client-id"
	FlagSourceNode           = "source-node"
)

// GetTxCmd returns the transaction commands for IBC Clients
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcTxCmd := &cobra.Command{
		Use:                        "client",
		Short:                      "Client transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcTxCmd.AddCommand(cli.PostCommands(
		GetCmdCreateClient(cdc),
		GetCmdUpdateClient(cdc),
	)...)

	return ibcTxCmd
}

// GetCmdCreateClient defines the command to create a new IBC Client as defined
// in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func GetCmdCreateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create new client with a consensus state",
		Long: strings.TrimSpace(`create new client with a specified identifier and consensus state:

		$ <appcli> tx ibc client create $CLIENTID ./state.json --from node0 --home ../node0/<app>cli --chain-id $CID
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var state exported.ConsensusState
			if err := cdc.UnmarshalJSON(contents, &state); err != nil {
				return err
			}

			msg := types.MsgCreateClient{
				ClientID:       args[0],
				ConsensusState: state,
				Signer:         cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// GetCmdUpdateClient defines the command to update a client as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#update
func GetCmdUpdateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update existing client with a header",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var header exported.Header
			if err := cdc.UnmarshalJSON(contents, &header); err != nil {
				return err
			}

			if err := header.ValidateBasic(txBldr.ChainID()); err != nil {
				return sdk.ErrInternal(err.Error()) // TODO: create error on types
			}

			msg := types.MsgUpdateClient{
				ClientID: args[0],
				Header:   header,
				Signer:   cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
