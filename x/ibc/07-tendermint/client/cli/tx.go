package cli

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	ics23 "github.com/confio/ics23/go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/light"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

const (
	flagTrustLevel = "trust-level"
	flagProofSpecs = "proof-specs"
)

// NewCreateClientCmd defines the command to create a new IBC Client as defined
// in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func NewCreateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [client-id] [path/to/consensus_state.json] [trusting_period] [unbonding_period] [max_clock_drift]",
		Short: "create new tendermint client",
		Long: `Create a new tendermint IBC client. 
  - 'trust-level' flag can be a fraction (eg: '1/3') or 'default'
  - 'proof-specs' flag can be JSON input, a path to a .json file or 'default'`,
		Example: fmt.Sprintf("%s tx ibc %s create [client-id] [path/to/consensus_state.json] [trusting_period] [unbonding_period] [max_clock_drift] --trust-level default --proof-specs [path/to/proof-specs.json] --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, types.SubModuleName),
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)
			legacyAmino := codec.New()

			var header *types.Header
			if err := cdc.UnmarshalJSON([]byte(args[1]), header); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided for consensus header")
				}
				if err := cdc.UnmarshalJSON(contents, header); err != nil {
					return errors.Wrap(err, "error unmarshalling consensus header file")
				}
			}

			var (
				trustLevel types.Fraction
				specs      []*ics23.ProofSpec
			)

			lvl, _ := cmd.Flags().GetString(flagTrustLevel)

			if lvl == "default" {
				trustLevel = types.NewFractionFromTm(light.DefaultTrustLevel)
			} else {
				trustLevel, err = parseFraction(lvl)
				if err != nil {
					return err
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

			maxClockDrift, err := time.ParseDuration(args[4])
			if err != nil {
				return err
			}

			spc, _ := cmd.Flags().GetString(flagProofSpecs)
			if spc == "default" {
				specs = commitmenttypes.GetSDKSpecs()
				// TODO migrate to use JSONMarshaler (implement MarshalJSONArray
				// or wrap lists of proto.Message in some other message)
			} else if err := legacyAmino.UnmarshalJSON([]byte(spc), &specs); err != nil {
				// check for file path if JSON input not provided
				contents, err := ioutil.ReadFile(spc)
				if err != nil {
					return errors.New("neither JSON input nor path to .json file was provided for proof specs flag")
				}
				// TODO migrate to use JSONMarshaler (implement MarshalJSONArray
				// or wrap lists of proto.Message in some other message)
				if err := legacyAmino.UnmarshalJSON(contents, &specs); err != nil {
					return errors.Wrap(err, "error unmarshalling proof specs file")
				}
			}

			msg := types.NewMsgCreateClient(
				clientID, header, trustLevel, trustingPeriod, ubdPeriod, maxClockDrift, specs, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagTrustLevel, "default", "light client trust level fraction for header updates")
	cmd.Flags().String(flagProofSpecs, "default", "proof specs format to be used for verification")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUpdateClientCmd defines the command to update a client as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#update
func NewUpdateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [client-id] [path/to/header.json]",
		Short: "update existing client with a header",
		Long:  "update existing tendermint client with a tendermint header",
		Example: fmt.Sprintf(
			"$ %s tx ibc %s update [client-id] [path/to/header.json] --from node0 --home ../node0/<app>cli --chain-id $CID",
			version.AppName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(2),
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

			msg := types.NewMsgUpdateClient(clientID, header, clientCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewSubmitMisbehaviourCmd defines the command to submit a misbehaviour to invalidate
// previous state roots and prevent future updates as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#misbehaviour
func NewSubmitMisbehaviourCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "misbehaviour [path/to/evidence.json]",
		Short: "submit a client misbehaviour",
		Long:  "submit a client misbehaviour to invalidate to invalidate previous state roots and prevent future updates",
		Example: fmt.Sprintf(
			"$ %s tx ibc %s misbehaviour [path/to/evidence.json] --from node0 --home ../node0/<app>cli --chain-id $CID",
			version.AppName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(1),
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

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func parseFraction(fraction string) (types.Fraction, error) {
	fr := strings.Split(fraction, "/")
	if len(fr) != 2 || fr[0] == fraction {
		return types.Fraction{}, fmt.Errorf("fraction must have format 'numerator/denominator' got %s", fraction)
	}

	numerator, err := strconv.ParseInt(fr[0], 10, 64)
	if err != nil {
		return types.Fraction{}, fmt.Errorf("invalid trust-level numerator: %w", err)
	}

	denominator, err := strconv.ParseInt(fr[1], 10, 64)
	if err != nil {
		return types.Fraction{}, fmt.Errorf("invalid trust-level denominator: %w", err)
	}

	return types.Fraction{
		Numerator:   numerator,
		Denominator: denominator,
	}, nil

}
