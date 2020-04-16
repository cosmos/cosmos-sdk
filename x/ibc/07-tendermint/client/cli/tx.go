package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// NewCreateClientTxCmd defines the command to create a new IBC Client as defined
// in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func NewCreateClientTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [client-id] [path/to/consensus_state.json] [trusting_period] [unbonding_period]",
		Short: "create new client with a consensus state",
		Long: strings.TrimSpace(fmt.Sprintf(`create new client with a specified identifier and consensus state:

Example:
$ %s tx ibc client create [client-id] [path/to/consensus_state.json] [trusting_period] [unbonding_period] --from node0 --home ../node0/<app>cli --chain-id $CID
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)
			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m).WithBroadcastMode(flags.BroadcastBlock)

			clientID := args[0]

			var header ibctmtypes.Header
			if err := m.UnmarshalJSON([]byte(args[1]), &header); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := m.UnmarshalJSON(contents, &header); err != nil {
					return errors.Wrap(err, "error unmarshalling consensus header file")
				}
			}

			trustingPeriod, err := time.ParseDuration(args[2])
			if err != nil {
				return err
			}

			ubdPeriod, err := time.ParseDuration(args[3])
			if err != nil {
				return err
			}

			msg := ibctmtypes.NewMsgCreateClient(clientID, header, trustingPeriod, ubdPeriod, cliCtx.GetFromAddress())

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}
	return cmd
}

// NewUpdateClientTxCmd defines the command to update a client as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#update
func NewUpdateClientTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
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
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)
			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m)

			clientID := args[0]

			var header ibctmtypes.Header
			if err := m.UnmarshalJSON([]byte(args[1]), &header); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := m.UnmarshalJSON(contents, &header); err != nil {
					return errors.Wrap(err, "error unmarshalling header file")
				}
			}

			msg := ibctmtypes.NewMsgUpdateClient(clientID, header, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}
	return cmd
}

// NewSubmitMisbehaviourTxCmd defines the command to submit a misbehaviour to invalidate
// previous state roots and prevent future updates as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#misbehaviour
func NewSubmitMisbehaviourTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "misbehaviour [path/to/evidence.json]",
		Short: "submit a client misbehaviour",
		Long: strings.TrimSpace(fmt.Sprintf(`submit a client misbehaviour to invalidate to invalidate previous state roots and prevent future updates:

Example:
$ %s tx ibc client misbehaviour [path/to/evidence.json] --from node0 --home ../node0/<app>cli --chain-id $CID
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)
			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m)

			var ev types.Evidence
			if err := m.UnmarshalJSON([]byte(args[0]), &ev); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[0])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := m.UnmarshalJSON(contents, &ev); err != nil {
					return errors.Wrap(err, "error unmarshalling evidence file")
				}
			}

			msg := ibctmtypes.NewMsgSubmitClientMisbehaviour(ev, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}
	return cmd
}
