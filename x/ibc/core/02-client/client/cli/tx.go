package cli

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// NewCreateClientCmd defines the command to create a new IBC light client.
func NewCreateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [path/to/client_state.json] [path/to/consensus_state.json]",
		Short: "create new IBC client",
		Long: `create a new IBC client with the specified client state and consensus state
	- ConsensusState json example: {"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A/3SXL2ONYaOkxpdR5P8tHTlSlPv1AwQwSFxKRee5JQW"},"diversifier":"diversifier","timestamp":"10"}`,
		Example: fmt.Sprintf("%s tx ibc %s create [path/to/client_state.json] [path/to/consensus_state] --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, types.SubModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			// attempt to unmarshal client state argument
			clientStateAny := &codectypes.Any{}
			if err := cdc.UnmarshalJSON([]byte(args[0]), clientStateAny); err != nil {

				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[0])
				if err != nil {
					return errors.Wrap(err, "neither JSON input nor path to .json file for client state were provided")
				}

				if err := cdc.UnmarshalJSON(contents, clientStateAny); err != nil {
					return errors.Wrap(err, "error unmarshalling client state file")
				}
			}

			// attempt to unmarshal consensus state argument
			consStateAny := &codectypes.Any{}
			if err := cdc.UnmarshalJSON([]byte(args[1]), consStateAny); err != nil {

				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.Wrap(err, "neither JSON input nor path to .json file for consensus state were provided")
				}

				if err := cdc.UnmarshalJSON(contents, consStateAny); err != nil {
					return errors.Wrap(err, "error unmarshalling consensus state file")
				}
			}

			msg := &types.MsgCreateClient{
				ClientState:    clientStateAny,
				ConsensusState: consStateAny,
				Signer:         clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// TODO: determine info to print out for client identifier created
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUpdateClientCmd defines the command to update an IBC client.
func NewUpdateClientCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "update [client-id] [path/to/header.json]",
		Short:   "update existing client with a header",
		Long:    "update existing client with a header",
		Example: fmt.Sprintf("%s tx ibc %s update [client-id] [path/to/header.json] --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, types.SubModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientID := args[0]

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var header exported.Header
			if err := cdc.UnmarshalJSON([]byte(args[1]), header); err != nil {

				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.Wrap(err, "neither JSON input nor path to .json file for header were provided")
				}

				if err := cdc.UnmarshalJSON(contents, header); err != nil {
					return errors.Wrap(err, "error unmarshalling header file")
				}
			}

			msg, err := types.NewMsgUpdateClient(clientID, header, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// NewSubmitMisbehaviourCmd defines the command to submit a misbehaviour to prevent
// future updates.
func NewSubmitMisbehaviourCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "misbehaviour [path/to/misbehaviour.json]",
		Short:   "submit a client misbehaviour",
		Long:    "submit a client misbehaviour to prevent future updates",
		Example: fmt.Sprintf("%s tx ibc %s misbehaviour [path/to/misbehaviour.json] --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, types.SubModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var misbehaviour exported.Misbehaviour
			if err := cdc.UnmarshalJSON([]byte(args[0]), misbehaviour); err != nil {

				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[0])
				if err != nil {
					return errors.Wrap(err, "neither JSON input nor path to .json file for misbehaviour were provided")
				}

				if err := cdc.UnmarshalJSON(contents, misbehaviour); err != nil {
					return errors.Wrap(err, "error unmarshalling misbehaviour file")
				}
			}

			msg, err := types.NewMsgSubmitMisbehaviour(misbehaviour.GetClientID(), misbehaviour, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
