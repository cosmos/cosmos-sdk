package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// GetTxCmd returns the transaction commands for IBC Clients
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics02ClientTxCmd := &cobra.Command{
		Use:                        "client",
		Short:                      "Client transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ics02ClientTxCmd.AddCommand(client.PostCommands(
		GetCmdCreateClient(cdc),
		GetCmdUpdateClient(cdc),
	)...)

	return ics02ClientTxCmd
}

// GetCmdCreateClient defines the command to create a new IBC Client as defined
// in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func GetCmdCreateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [client-id] [path/to/consensus_state.json]",
		Short: "create new client with a consensus state",
		Long: strings.TrimSpace(fmt.Sprintf(`create new client with a specified identifier and consensus state:

Example:
$ %s tx ibc client create [client-id] [path/to/consensus_state.json] --from node0 --home ../node0/<app>cli --chain-id $CID
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			clientID := args[0]

			var state exported.ConsensusState
			if err := cdc.UnmarshalJSON([]byte(args[1]), &state); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")

				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v", err)
				}
				if err := cdc.UnmarshalJSON(contents, &state); err != nil {
					return fmt.Errorf("error unmarshalling consensus state file: %v", err)
				}
			}

			msg := types.NewMsgCreateClient(
				clientID, state.ClientType().String(), state,
				cliCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
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
		Use:   "update [client-id] [path/to/header.json]",
		Short: "update existing client with a header",
		Long: strings.TrimSpace(fmt.Sprintf(`update existing client with a header:

Example:
$ %s tx ibc client update [client-id] [path/to/header.json] --from node0 --home ../node0/<app>cli --chain-id $CID
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			clientID := args[0]

			var header exported.Header
			if err := cdc.UnmarshalJSON([]byte(args[1]), &header); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")

				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v", err)
				}
				if err := cdc.UnmarshalJSON(contents, &header); err != nil {
					return fmt.Errorf("error unmarshalling header file: %v", err)
				}
			}

			msg := types.NewMsgUpdateClient(clientID, header, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// GetCmdSubmitMisbehaviour defines the command to submit a misbehaviour to invalidate
// previous state roots and prevent future updates as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#misbehaviour
func GetCmdSubmitMisbehaviour(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "misbehaviour [client-it] [path/to/evidence.json]",
		Short: "submit a client misbehaviour",
		Long: strings.TrimSpace(fmt.Sprintf(`submit a client misbehaviour to invalidate to invalidate previous state roots and prevent future updates:

Example:
$ %s tx ibc client misbehaviour [client-id] [path/to/evidence.json] --from node0 --home ../node0/<app>cli --chain-id $CID
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			clientID := args[0]

			var evidence exported.Evidence
			if err := cdc.UnmarshalJSON([]byte(args[1]), &evidence); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")

				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v", err)
				}
				if err := cdc.UnmarshalJSON(contents, &evidence); err != nil {
					return fmt.Errorf("error unmarshalling evidence file: %v", err)
				}
			}

			msg := types.NewMsgSubmitMisbehaviour(clientID, evidence, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
