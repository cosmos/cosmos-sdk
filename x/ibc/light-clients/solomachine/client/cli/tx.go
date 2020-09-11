package cli

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

const (
	flagAllowUpdateAfterProposal = "allow_update_after_proposal"
)

// NewCreateClientCmd defines the command to create a new solo machine client.
func NewCreateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create [client-id] [sequence] [path/to/consensus_state.json]",
		Short:   "create new solo machine client",
		Long:    "create a new solo machine client with the specified identifier and consensus state",
		Example: fmt.Sprintf("%s tx ibc %s create [client-id] [sequence] [path/to/consensus_state.json] --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, types.SubModuleName),
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]

			sequence, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var consensusState *types.ConsensusState
			if err := cdc.UnmarshalJSON([]byte(args[2]), consensusState); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := cdc.UnmarshalJSON(contents, consensusState); err != nil {
					return errors.Wrap(err, "error unmarshalling consensus state file")
				}
			}

			allowUpdateAfterProposal, _ := cmd.Flags().GetBool(flagAllowUpdateAfterProposal)

			clientState := types.NewClientState(sequence, consensusState, allowUpdateAfterProposal)
			msg, err := clienttypes.NewMsgCreateClient(clientID, clientState, consensusState, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(flagAllowUpdateAfterProposal, false, "allow governance proposal to update client")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
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

			msg, err := clienttypes.NewMsgUpdateClient(clientID, header, clientCtx.GetFromAddress())
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
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var m *types.Misbehaviour
			if err := cdc.UnmarshalJSON([]byte(args[0]), m); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[0])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := cdc.UnmarshalJSON(contents, m); err != nil {
					return errors.Wrap(err, "error unmarshalling misbehaviour file")
				}
			}

			msg, err := clienttypes.NewMsgSubmitMisbehaviour(m.ClientId, m, clientCtx.GetFromAddress())
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
