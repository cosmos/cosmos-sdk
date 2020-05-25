package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmmath "github.com/tendermint/tendermint/libs/math"
	lite "github.com/tendermint/tendermint/lite2"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

const flagTrustLevel = "trust-level"

// GetCmdCreateClient defines the command to create a new IBC Client as defined
// in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func GetCmdCreateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [client-id] [path/to/consensus_state.json] [trusting_period] [unbonding_period] [max_clock_drift]",
		Short: "create new tendermint client",
		Long:  "create new tendermint client. Trust level can be a fraction (eg: '1/3') or 'default'",
		Example: fmt.Sprintf(
			"%s tx ibc %s create [client-id] [path/to/consensus_state.json] [trusting_period] [unbonding_period] [max_clock_drift] --trust-level default --from node0 --home ../node0/<app>cli --chain-id $CID",
			version.ClientName, ibctmtypes.SubModuleName),
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			clientID := args[0]

			var header ibctmtypes.Header
			if err := cdc.UnmarshalJSON([]byte(args[1]), &header); err != nil {
				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return errors.New("neither JSON input nor path to .json file were provided")
				}
				if err := cdc.UnmarshalJSON(contents, &header); err != nil {
					return errors.Wrap(err, "error unmarshalling consensus header file")
				}
			}

			var (
				trustLevel tmmath.Fraction
				err        error
			)

			lvl := viper.GetString(flagTrustLevel)

			if lvl == "default" {
				trustLevel = lite.DefaultTrustLevel
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

			msg := ibctmtypes.NewMsgCreateClient(
				clientID, header, trustLevel, trustingPeriod, ubdPeriod, maxClockDrift, cliCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagTrustLevel, "default", "light client trust level fraction for header updates")
	return cmd
}

// GetCmdUpdateClient defines the command to update a client as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#update
func GetCmdUpdateClient(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "update [client-id] [path/to/header.json]",
		Short: "update existing client with a header",
		Long:  "update existing tendermint client with a tendermint header",
		Example: fmt.Sprintf(
			"$ %s tx ibc %s update [client-id] [path/to/header.json] --from node0 --home ../node0/<app>cli --chain-id $CID",
			version.ClientName, ibctmtypes.SubModuleName,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			clientID := args[0]

			var header ibctmtypes.Header
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

			msg := ibctmtypes.NewMsgUpdateClient(clientID, header, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSubmitMisbehaviour defines the command to submit a misbehaviour to invalidate
// previous state roots and prevent future updates as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#misbehaviour
func GetCmdSubmitMisbehaviour(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "misbehaviour [path/to/evidence.json]",
		Short: "submit a client misbehaviour",
		Long:  "submit a client misbehaviour to invalidate to invalidate previous state roots and prevent future updates",
		Example: fmt.Sprintf(
			"$ %s tx ibc %s misbehaviour [path/to/evidence.json] --from node0 --home ../node0/<app>cli --chain-id $CID",
			version.ClientName, ibctmtypes.SubModuleName,
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

			msg := ibctmtypes.NewMsgSubmitClientMisbehaviour(ev, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func parseFraction(fraction string) (tmmath.Fraction, error) {
	fr := strings.Split(fraction, "/")
	if len(fr) != 2 || fr[0] == fraction {
		return tmmath.Fraction{}, fmt.Errorf("fraction must have format 'numerator/denominator' got %s", fraction)
	}

	numerator, err := strconv.ParseInt(fr[0], 10, 64)
	if err != nil {
		return tmmath.Fraction{}, fmt.Errorf("invalid trust-level numerator: %w", err)
	}

	denominator, err := strconv.ParseInt(fr[1], 10, 64)
	if err != nil {
		return tmmath.Fraction{}, fmt.Errorf("invalid trust-level denominator: %w", err)
	}

	return tmmath.Fraction{
		Numerator:   numerator,
		Denominator: denominator,
	}, nil
}
