package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	ibcsmtypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
)

// GetCmdCreateClient defines the command to create a new IBC Client.
func GetCmdCreateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [client-id] [path/to/consensus_state.json]",
		Short: "create new client with a consensus state",
		Long: strings.TrimSpace(fmt.Sprintf(` create new client with specified identifier and consensus state:

Example:
$ %s tx ibc client create [client-id] [path/to/consensus_state.json] --from node0 --home ../node0/<app>cli --chain-id $CID
	`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flag.BroadcastBlock)

			clientID := args[0]

			var consensusState ibcsmtypes.ConsensusState
			if err := cdc.UnmarshalJSON([]byte(args[1]), &consensusState); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := cdc.UnmarshalJSON(contents, &consensusState); err != nil {
					return errors.Wrap(err, "error unmarshalling consensus header file")
				}
			}

			msg := ibcsmtypes.NewMsgCreateClient(clientID, consensusState)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// GetCmdUpdateClient defines the command to update a client.
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
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			clientID := args[0]

			var header ibcsmtypes.Header
			if err := cdc.UnmarshalJSON([]byte(args[1]), &header); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := cdc.UnmarshalJSON(contents, &header); err != nil {
					return errors.Wrap(err, "error unmarshalling header file")
				}
			}

			msg := ibcsmtypes.NewMsgUpdateClient(clientID, header)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// GetCmdSubmitMisbehaviour defines the command to submit a misbehaviour to prevent
// future updates as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#misbehaviour
func GetCmdSubmitMisbehaviour(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "misbehaviour [path/to/evidence.json]",
		Short: "submit a client misbehaviour",
		Long: strings.TrimSpace(fmt.Sprintf(`submit a client misbehaviour to prevent future updates:

Example:
$ %s tx ibc client misbehaviour [path/to/evidence.json] --from node0 --home ../node0/<app>cli --chain-id $CID
	`, version.ClientName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			var ev evidenceexported.Evidence
			if err := cdc.UnmarshalJSON([]byte(args[0]), &ev); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[0])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := cdc.UnmarshalJSON(contents, &ev); err != nil {
					return errors.Wrap(err, "error unmarshalling evidence file")
				}
			}

			msg := ibcsmtypes.NewMsgSubmitClientMisbehaviour(ev, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
