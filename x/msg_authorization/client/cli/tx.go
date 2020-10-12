package cli

import (
	"errors"
	"io/ioutil"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string) *cobra.Command {
	AuthorizationTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Authorization transactions subcommands",
		Long:                       "Authorize and revoke access to execute transactions on behalf of your address",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	AuthorizationTxCmd.AddCommand(
		GetCmdGrantAuthorization(storeKey),
		GetCmdRevokeAuthorization(storeKey),
		GetCmdSendAs(storeKey),
	)

	return AuthorizationTxCmd
}

func GetCmdGrantAuthorization(storeKey string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant [grantee_address] [type] [authorization] --from=[granter]",
		Short: "Grant authorization to an address",
		Long:  "Grant authorization to an address to execute a transaction on your behalf",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msgType := args[1]

			bz, err := ioutil.ReadFile(args[2])
			if err != nil {
				return err
			}

			var authorization types.Authorization
			switch msgType {
			case (types.SendAuthorization{}.MsgType()):
				var sendAuth types.SendAuthorization
				err := clientCtx.JSONMarshaler.UnmarshalJSON(bz, &sendAuth)
				if err != nil {
					return err
				}
				authorization = &sendAuth
			case (types.GenericAuthorization{}.MsgType()):
				var genAuth types.GenericAuthorization
				err := clientCtx.JSONMarshaler.UnmarshalJSON(bz, &genAuth)
				if err != nil {
					return err
				}
				authorization = &genAuth
			default:
				return errors.New("invalid authorization type")
			}

			period := time.Unix(viper.GetInt64(FlagExpiration), 0)

			msg, err := types.NewMsgGrantAuthorization(clientCtx.GetFromAddress(), grantee, authorization, period)
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
	cmd.Flags().Int64(FlagExpiration, int64(3600*24*365), "The second unit of time duration which the authorization is active for the user; Default is a year")
	return cmd
}

func GetCmdRevokeAuthorization(storeKey string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [grantee_address] [msg_type] --from=[granter_address]",
		Short: "revoke authorization",
		Long:  "revoke authorization from an address for a transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
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
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetCmdSendAs(storeKey string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-as [grantee] [msg_tx_json] --from [grantee]",
		Short: "execute tx on behalf of granter account",
		Long:  "execute tx on behalf of granter account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			if offline, _ := cmd.Flags().GetBool(flags.FlagOffline); offline {
				return errors.New("cannot broadcast tx during offline mode")
			}

			stdTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgExecAuthorized(grantee, stdTx.GetMsgs())

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
