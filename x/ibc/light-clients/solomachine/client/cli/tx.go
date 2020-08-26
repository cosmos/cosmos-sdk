package cli

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

// NewCreateClientCmd defines the command to create a new solo machine client.
func NewCreateClientCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "create [client-id] [path/to/consensus_state.json]",
		Short:   "create new solo machine client",
		Long:    "create a new solo machine client with the specified identifier and consensus state",
		Example: fmt.Sprintf("%s tx ibc %s create [client-id] [path/to/consensus_state.json] --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, types.SubModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var consensusState *types.ConsensusState
			if err := cdc.UnmarshalJSON([]byte(args[1]), consensusState); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := cdc.UnmarshalJSON(contents, consensusState); err != nil {
					return errors.Wrap(err, "error unmarshalling consensus state file")
				}
			}

			msg := types.NewMsgCreateClient(clientID, consensusState)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// NewUpdateClientCmd defines the command to update a solo machine client.
func NewUpdateClientCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "update [client-id] [path/to/header.json]",
		Short:   "update existing client with a header",
		Long:    "update existing client with a solo machine header",
		Example: fmt.Sprintf("%s tx ibc %s update [client-id] [path/to/header.json] --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, types.SubModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var header *types.Header
			if err := cdc.UnmarshalJSON([]byte(args[1]), header); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := cdc.UnmarshalJSON(contents, header); err != nil {
					return errors.Wrap(err, "error unmarshalling header file")
				}
			}

			msg := types.NewMsgUpdateClient(clientID, header)
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
		Use:     "misbehaviour [path/to/evidence.json]",
		Short:   "submit a client misbehaviour",
		Long:    "submit a client misbehaviour to prevent future updates",
		Example: fmt.Sprintf("%s tx ibc %s misbehaviour [path/to/evidence.json] --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, types.SubModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var ev *types.Evidence
			if err := cdc.UnmarshalJSON([]byte(args[0]), ev); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[0])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := cdc.UnmarshalJSON(contents, ev); err != nil {
					return errors.Wrap(err, "error unmarshalling evidence file")
				}
			}

			msg := types.NewMsgSubmitClientMisbehaviour(ev, clientCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
