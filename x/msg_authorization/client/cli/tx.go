package cli

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/client/flags"

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
		Use:   "grant [grantee_address] [authorization] [limit]",
		Short: "Grant authorization to an address",
		Long:  "Grant authorization to an address to execute a transaction on your behalf",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			granter := clientCtx.GetFromAddress()
			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msgType  := args[1]

			var authorization types.Authorization
			var (
				_ types.Authorization = &types.SendAuthorization{}
			)
			if msgType == (types.SendAuthorization{}.MsgType()) {
				limit, err := sdk.ParseCoins(args[2])
				if err != nil {
					return nil
				}
				authorization = types.NewSendAuthorization(limit)
			} else if msgType == (types.GenericAuthorization{}.MsgType()) {
				limit, err := sdk.ParseCoins(args[2])
				if err != nil {
					return nil
				}
				authorization = types.NewSendAuthorization(limit)
				// TODO replace with generic authorization
				// authorization = types.NewGenericAuthorization(msgType)
			} else {
				return nil
			}

			expirationString := viper.GetString(FlagExpiration)
			expiration, err := time.Parse(time.RFC3339, expirationString)
			if err != nil {
				return err
			}

			msg, err  := types.NewMsgGrantAuthorization(granter, grantee, authorization, expiration)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)

		},
	}
	cmd.Flags().String(FlagExpiration, "9999-12-31T23:59:59.52Z", "The time upto which the authorization is active for the user")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetCmdRevokeAuthorization(storeKey string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [grantee_address] [msg_type]",
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

			return nil
			// inBuf := bufio.NewReader(cmd.InOrStdin())
			// txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			// cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			// grantee := cliCtx.FromAddress

			// var stdTx auth.StdTx
			// bz, err := ioutil.ReadFile(args[1])
			// if err != nil {
			// 	return err
			// }

			// err = cdc.UnmarshalJSON(bz, &stdTx)
			// if err != nil {
			// 	return err
			// }

			// msg := types.NewMsgExecAuthorized(grantee, stdTx.Msgs)

			// if err := msg.ValidateBasic(); err != nil {
			// 	return err
			// }

			// return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
