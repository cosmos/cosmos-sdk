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
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
)

const (
	flagTrustLevel                   = "trust-level"
	flagProofSpecs                   = "proof-specs"
	flagUpgradePath                  = "upgrade-path"
	flagAllowUpdateAfterExpiry       = "allow_update_after_expiry"
	flagAllowUpdateAfterMisbehaviour = "allow_update_after_misbehaviour"
)

// NewCreateClientCmd defines the command to create a new IBC Client as defined
// in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func NewCreateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [path/to/consensus_state.json] [trusting_period] [unbonding_period] [max_clock_drift]",
		Short: "create new tendermint client",
		Long: `Create a new tendermint IBC client.
  - 'trust-level' flag can be a fraction (eg: '1/3') or 'default'
  - 'proof-specs' flag can be JSON input, a path to a .json file or 'default'
  - 'upgrade-path' flag is a string specifying the upgrade path for this chain where a future upgraded client will be stored. The path is a comma-separated list representing the keys in order of the keyPath to the committed upgraded client.
  e.g. 'upgrade/upgradedClient'`,
		Example: fmt.Sprintf("%s tx ibc %s create [path/to/consensus_state.json] [trusting_period] [unbonding_period] [max_clock_drift] --trust-level default --consensus-params [path/to/consensus-params.json] --proof-specs [path/to/proof-specs.json] --upgrade-path upgrade/upgradedClient --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, types.SubModuleName),
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)
			legacyAmino := codec.NewLegacyAmino()

			var header *types.Header
			if err := cdc.UnmarshalJSON([]byte(args[0]), header); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[0])
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

			trustingPeriod, err := time.ParseDuration(args[1])
			if err != nil {
				return err
			}

			ubdPeriod, err := time.ParseDuration(args[2])
			if err != nil {
				return err
			}

			maxClockDrift, err := time.ParseDuration(args[3])
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

			allowUpdateAfterExpiry, _ := cmd.Flags().GetBool(flagAllowUpdateAfterExpiry)
			allowUpdateAfterMisbehaviour, _ := cmd.Flags().GetBool(flagAllowUpdateAfterMisbehaviour)

			upgradePathStr, _ := cmd.Flags().GetString(flagUpgradePath)
			upgradePath := strings.Split(upgradePathStr, ",")

			// validate header
			if err := header.ValidateBasic(); err != nil {
				return err
			}

			height := header.GetHeight().(clienttypes.Height)

			clientState := types.NewClientState(
				header.GetHeader().GetChainID(), trustLevel, trustingPeriod, ubdPeriod, maxClockDrift,
				height, specs, upgradePath, allowUpdateAfterExpiry, allowUpdateAfterMisbehaviour,
			)

			consensusState := header.ConsensusState()

			msg, err := clienttypes.NewMsgCreateClient(
				clientState, consensusState, clientCtx.GetFromAddress(),
			)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagTrustLevel, "default", "light client trust level fraction for header updates")
	cmd.Flags().String(flagProofSpecs, "default", "proof specs format to be used for verification")
	cmd.Flags().Bool(flagAllowUpdateAfterExpiry, false, "allow governance proposal to update client after expiry")
	cmd.Flags().Bool(flagAllowUpdateAfterMisbehaviour, false, "allow governance proposal to update client after misbehaviour")
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
			clientCtx, err := client.GetClientTxContext(cmd)
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

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewSubmitMisbehaviourCmd defines the command to submit a misbehaviour to invalidate
// previous state roots and prevent future updates as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#misbehaviour
func NewSubmitMisbehaviourCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "misbehaviour [path/to/misbehaviour.json]",
		Short: "submit a client misbehaviour",
		Long:  "submit a client misbehaviour to invalidate to invalidate previous state roots and prevent future updates",
		Example: fmt.Sprintf(
			"$ %s tx ibc %s misbehaviour [path/to/misbehaviour.json] --from node0 --home ../node0/<app>cli --chain-id $CID",
			version.AppName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
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

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func parseFraction(fraction string) (types.Fraction, error) {
	fr := strings.Split(fraction, "/")
	if len(fr) != 2 || fr[0] == fraction {
		return types.Fraction{}, fmt.Errorf("fraction must have format 'numerator/denominator' got %s", fraction)
	}

	numerator, err := strconv.ParseUint(fr[0], 10, 64)
	if err != nil {
		return types.Fraction{}, fmt.Errorf("invalid trust-level numerator: %w", err)
	}

	denominator, err := strconv.ParseUint(fr[1], 10, 64)
	if err != nil {
		return types.Fraction{}, fmt.Errorf("invalid trust-level denominator: %w", err)
	}

	return types.Fraction{
		Numerator:   numerator,
		Denominator: denominator,
	}, nil

}
