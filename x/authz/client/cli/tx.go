package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	authnclient "github.com/cosmos/cosmos-sdk/x/authn/client"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
)

const FlagSpendLimit = "spend-limit"
const FlagMsgType = "msg-type"
const FlagExpiration = "expiration"

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	AuthorizationTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Authorization transactions subcommands",
		Long:                       "Authorize and revoke access to execute transactions on behalf of your address",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	AuthorizationTxCmd.AddCommand(
		NewCmdGrantAuthorization(),
		NewCmdRevokeAuthorization(),
		NewCmdExecAuthorization(),
	)

	return AuthorizationTxCmd
}

func NewCmdGrantAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant <grantee> <authorization_type=\"send\"|\"generic\"> --from <granter>",
		Short: "Grant authorization to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Grant authorization to an address to execute a transaction on your behalf:

Examples:
 $ %s tx %s grant cosmos1skjw.. send %s --spend-limit=1000stake --from=cosmos1skl..
 $ %s tx %s grant cosmos1skjw.. generic --msg-type=/cosmos.gov.v1beta1.Msg/Vote --from=cosmos1sk..
	`, version.AppName, types.ModuleName, types.SendAuthorization{}.MethodName(), version.AppName, types.ModuleName),
		),
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			var authorization types.Authorization
			switch args[1] {
			case "send":
				limit, err := cmd.Flags().GetString(FlagSpendLimit)
				if err != nil {
					return err
				}

				spendLimit, err := sdk.ParseCoinsNormalized(limit)
				if err != nil {
					return err
				}

				if !spendLimit.IsAllPositive() {
					return fmt.Errorf("spend-limit should be greater than zero")
				}

				authorization = &types.SendAuthorization{
					SpendLimit: spendLimit,
				}
			case "generic":
				msgType, err := cmd.Flags().GetString(FlagMsgType)
				if err != nil {
					return err
				}

				authorization = types.NewGenericAuthorization(msgType)
			default:
				return fmt.Errorf("invalid authorization type, %s", args[1])
			}

			exp, err := cmd.Flags().GetInt64(FlagExpiration)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgGrantAuthorization(clientCtx.GetFromAddress(), grantee, authorization, time.Unix(exp, 0))
			if err != nil {
				return err
			}

			svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
			authzMsgClient := types.NewMsgClient(svcMsgClientConn)
			_, err = authzMsgClient.GrantAuthorization(context.Background(), msg)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), svcMsgClientConn.GetMsgs()...)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagMsgType, "", "The Msg method name for which we are creating a GenericAuthorization")
	cmd.Flags().String(FlagSpendLimit, "", "SpendLimit for Send Authorization, an array of Coins allowed spend")
	cmd.Flags().Int64(FlagExpiration, time.Now().AddDate(1, 0, 0).Unix(), "The Unix timestamp. Default is one year.")
	return cmd
}

func NewCmdRevokeAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [grantee_address] [msg_type] --from=[granter_address]",
		Short: "revoke authorization",
		Long: strings.TrimSpace(
			fmt.Sprintf(`revoke authorization from a granter to a grantee:
Example:
 $ %s tx %s revoke cosmos1skj.. %s --from=cosmos1skj..
			`, version.AppName, types.ModuleName, types.SendAuthorization{}.MethodName()),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			granter := clientCtx.GetFromAddress()

			msgAuthorized := args[1]

			msg := types.NewMsgRevokeAuthorization(granter, grantee, msgAuthorized)

			svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
			authzMsgClient := types.NewMsgClient(svcMsgClientConn)
			_, err = authzMsgClient.RevokeAuthorization(context.Background(), &msg)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), svcMsgClientConn.GetMsgs()...)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCmdExecAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [msg_tx_json_file] --from [grantee]",
		Short: "execute tx on behalf of granter account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`execute tx on behalf of granter account:
Example:
 $ %s tx %s exec tx.json --from grantee
 $ %s tx bank send <granter> <recipient> --from <granter> --chain-id <chain-id> --generate-only > tx.json && %s tx %s exec tx.json --from grantee
			`, version.AppName, types.ModuleName, version.AppName, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			grantee := clientCtx.GetFromAddress()

			if offline, _ := cmd.Flags().GetBool(flags.FlagOffline); offline {
				return errors.New("cannot broadcast tx during offline mode")
			}

			theTx, err := authnclient.ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return err
			}
			msgs := theTx.GetMsgs()
			serviceMsgs := make([]sdk.ServiceMsg, len(msgs))
			for i, msg := range msgs {
				srvMsg, ok := msg.(sdk.ServiceMsg)
				if !ok {
					return fmt.Errorf("tx contains %T which is not a sdk.ServiceMsg", msg)
				}
				serviceMsgs[i] = srvMsg
			}

			msg := types.NewMsgExecAuthorized(grantee, serviceMsgs)
			svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
			authzMsgClient := types.NewMsgClient(svcMsgClientConn)
			_, err = authzMsgClient.ExecAuthorized(context.Background(), &msg)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), svcMsgClientConn.GetMsgs()...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
