package cli

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

const (
	flagVersionIdentifier = "version-identifier"
	flagVersionFeatures   = "version-features"
	flagDelayPeriod       = "delay-period"
)

// NewConnectionOpenInitCmd defines the command to initialize a connection on
// chain A with a given counterparty chain B
func NewConnectionOpenInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-init [client-id] [counterparty-client-id] [path/to/counterparty_prefix.json]",
		Short: "Initialize connection on chain A",
		Long: `Initialize a connection on chain A with a given counterparty chain B.
	- 'version-identifier' flag can be a single pre-selected version identifier to be used in the handshake.
	- 'version-features' flag can be a list of features separated by commas to accompany the version identifier.`,
		Example: fmt.Sprintf(
			"%s tx %s %s open-init [client-id] [counterparty-client-id] [path/to/counterparty_prefix.json] --version-identifier=\"1.0\" --version-features=\"ORDER_UNORDERED\" --delay-period=500",
			version.AppName, host.ModuleName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientID := args[0]
			counterpartyClientID := args[1]

			counterpartyPrefix, err := utils.ParsePrefix(clientCtx.LegacyAmino, args[2])
			if err != nil {
				return err
			}

			var version *types.Version
			versionIdentifier, _ := cmd.Flags().GetString(flagVersionIdentifier)

			if versionIdentifier != "" {
				var features []string

				versionFeatures, _ := cmd.Flags().GetString(flagVersionFeatures)
				if versionFeatures != "" {
					features = strings.Split(versionFeatures, ",")
				}

				version = types.NewVersion(versionIdentifier, features)
			}

			delayPeriod, err := cmd.Flags().GetUint64(flagDelayPeriod)
			if err != nil {
				return err
			}

			msg := types.NewMsgConnectionOpenInit(
				clientID, counterpartyClientID,
				counterpartyPrefix, version, delayPeriod, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// NOTE: we should use empty default values since the user may not want to select a version
	// at this step in the handshake.
	cmd.Flags().String(flagVersionIdentifier, "", "version identifier to be used in the connection handshake version negotiation")
	cmd.Flags().String(flagVersionFeatures, "", "version features list separated by commas without spaces. The features must function with the version identifier.")
	cmd.Flags().Uint64(flagDelayPeriod, 0, "delay period that must pass before packet verification can pass against a consensus state")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewConnectionOpenTryCmd defines the command to relay a try open a connection on
// chain B
func NewConnectionOpenTryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: strings.TrimSpace(`open-try [connection-id] [client-id]
[counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json] [path/to/client_state.json]
[path/to/counterparty_version1.json,path/to/counterparty_version2.json...] [consensus-height] [proof-height] [path/to/proof_init.json] [path/to/proof_client.json] [path/to/proof_consensus.json]`),
		Short: "initiate connection handshake between two chains",
		Long:  "Initialize a connection on chain A with a given counterparty chain B. Provide counterparty versions separated by commas",
		Example: fmt.Sprintf(
			`%s tx %s %s open-try connection-id] [client-id] \
[counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json] [path/to/client_state.json]\
[counterparty-versions] [consensus-height] [proof-height] [path/to/proof_init.json] [path/to/proof_client.json] [path/to/proof_consensus.json]`,
			version.AppName, host.ModuleName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(12),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			connectionID := args[0]
			clientID := args[1]
			counterpartyConnectionID := args[2]
			counterpartyClientID := args[3]

			counterpartyPrefix, err := utils.ParsePrefix(clientCtx.LegacyAmino, args[4])
			if err != nil {
				return err
			}

			counterpartyClient, err := utils.ParseClientState(clientCtx.LegacyAmino, args[5])
			if err != nil {
				return err
			}

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			versionsStr := strings.Split(args[6], ",")
			counterpartyVersions := make([]*types.Version, len(versionsStr))

			for _, ver := range versionsStr {

				// attempt to unmarshal version
				version := &types.Version{}
				if err := cdc.UnmarshalJSON([]byte(ver), version); err != nil {

					// check for file path if JSON input is not provided
					contents, err := ioutil.ReadFile(ver)
					if err != nil {
						return errors.Wrap(err, "neither JSON input nor path to .json file for version were provided")
					}

					if err := cdc.UnmarshalJSON(contents, version); err != nil {
						return errors.Wrap(err, "error unmarshalling version file")
					}
				}
			}

			consensusHeight, err := clienttypes.ParseHeight(args[7])
			if err != nil {
				return err
			}
			proofHeight, err := clienttypes.ParseHeight(args[8])
			if err != nil {
				return err
			}

			proofInit, err := utils.ParseProof(clientCtx.LegacyAmino, args[9])
			if err != nil {
				return err
			}

			proofClient, err := utils.ParseProof(clientCtx.LegacyAmino, args[10])
			if err != nil {
				return err
			}

			proofConsensus, err := utils.ParseProof(clientCtx.LegacyAmino, args[11])
			if err != nil {
				return err
			}

			delayPeriod, err := cmd.Flags().GetUint64(flagDelayPeriod)
			if err != nil {
				return err
			}

			msg := types.NewMsgConnectionOpenTry(
				connectionID, clientID, counterpartyConnectionID, counterpartyClientID,
				counterpartyClient, counterpartyPrefix, counterpartyVersions, delayPeriod,
				proofInit, proofClient, proofConsensus, proofHeight,
				consensusHeight, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Uint64(flagDelayPeriod, 0, "delay period that must pass before packet verification can pass against a consensus state")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewConnectionOpenAckCmd defines the command to relay the acceptance of a
// connection open attempt from chain B to chain A
func NewConnectionOpenAckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: `open-ack [connection-id] [counterparty-connection-id] [path/to/client_state.json] [consensus-height] [proof-height]
		[path/to/proof_try.json] [path/to/proof_client.json] [path/to/proof_consensus.json] [version]`,
		Short: "relay the acceptance of a connection open attempt",
		Long:  "Relay the acceptance of a connection open attempt from chain B to chain A",
		Example: fmt.Sprintf(
			`%s tx %s %s open-ack [connection-id] [counterparty-connection-id] [path/to/client_state.json] [consensus-height] [proof-height]
			[path/to/proof_try.json] [path/to/proof_client.json] [path/to/proof_consensus.json] [version]`,
			version.AppName, host.ModuleName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(9),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			connectionID := args[0]
			counterpartyConnectionID := args[1]

			counterpartyClient, err := utils.ParseClientState(clientCtx.LegacyAmino, args[2])
			if err != nil {
				return err
			}

			consensusHeight, err := clienttypes.ParseHeight(args[3])
			if err != nil {
				return err
			}
			proofHeight, err := clienttypes.ParseHeight(args[4])
			if err != nil {
				return err
			}

			proofTry, err := utils.ParseProof(clientCtx.LegacyAmino, args[5])
			if err != nil {
				return err
			}

			proofClient, err := utils.ParseProof(clientCtx.LegacyAmino, args[6])
			if err != nil {
				return err
			}

			proofConsensus, err := utils.ParseProof(clientCtx.LegacyAmino, args[7])
			if err != nil {
				return err
			}

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			// attempt to unmarshal version
			version := &types.Version{}
			if err := cdc.UnmarshalJSON([]byte(args[8]), version); err != nil {

				// check for file path if JSON input is not provided
				contents, err := ioutil.ReadFile(args[8])
				if err != nil {
					return errors.Wrap(err, "neither JSON input nor path to .json file for version were provided")
				}

				if err := cdc.UnmarshalJSON(contents, version); err != nil {
					return errors.Wrap(err, "error unmarshalling version file")
				}
			}

			msg := types.NewMsgConnectionOpenAck(
				connectionID, counterpartyConnectionID, counterpartyClient, proofTry, proofClient, proofConsensus, proofHeight,
				consensusHeight, version, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewConnectionOpenConfirmCmd defines the command to initialize a connection on
// chain A with a given counterparty chain B
func NewConnectionOpenConfirmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-confirm [connection-id] [proof-height] [path/to/proof_ack.json]",
		Short: "confirm to chain B that connection is open on chain A",
		Long:  "Confirm to chain B that connection is open on chain A",
		Example: fmt.Sprintf(
			"%s tx %s %s open-confirm [connection-id] [proof-height] [path/to/proof_ack.json]",
			version.AppName, host.ModuleName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			connectionID := args[0]
			proofHeight, err := clienttypes.ParseHeight(args[1])
			if err != nil {
				return err
			}

			proofAck, err := utils.ParseProof(clientCtx.LegacyAmino, args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgConnectionOpenConfirm(
				connectionID, proofAck, proofHeight, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
