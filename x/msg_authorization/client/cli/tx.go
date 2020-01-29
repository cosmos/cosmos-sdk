package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/exported"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	AuthorizationTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Authorization transactions subcommands",
		Long:                       "Authorize and revoke access to execute transactions on behalf of your address",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	AuthorizationTxCmd.AddCommand(flags.PostCommands(
		GetCmdGrantAuthorization(cdc),
		GetCmdRevokeAuthorization(cdc),
		GetCmdSendAs(cdc),
	)...)

	return AuthorizationTxCmd
}

func GetCmdGrantAuthorization(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant [grantee_address] [authorization] --from [granter_address_or_key]",
		Short: "Grant authorization to an address",
		Long:  "Grant authorization to an address to execute a transaction on your behalf",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			granter := cliCtx.FromAddress
			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			//
			//bz, err := ioutil.ReadFile(args[1])
			//if err != nil {
			//	return err
			//}

			bz := []byte(`{"type": "cosmos-sdk/SendAuthorization","value":{"spendlimit": [{"denom": "stake","amount": "50"}]}}`)

			var authorization exported.Authorization
			err = cdc.UnmarshalJSON(bz, &authorization)
			if err != nil {
				return err
			}
			fmt.Printf("this is authorization %v", authorization)
			expirationString := viper.GetString(FlagExpiration)
			expiration, err := time.Parse(time.RFC3339, expirationString)
			if err != nil {
				return err
			}

			msg := types.NewMsgGrantAuthorization(granter, grantee, authorization, expiration)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{msg})

		},
	}
	cmd.Flags().String(FlagExpiration, "9999-12-31T23:59:59.52Z", "The time upto which the authorization is active for the user")

	return cmd
}

func GetCmdRevokeAuthorization(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [grantee_address] [msg_type] --from [granter]",
		Short: "revoke authorization",
		Long:  "revoke authorization from an address for a transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			granter := cliCtx.FromAddress
			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var msgAuthorized sdk.Msg
			err = cdc.UnmarshalJSON(bz, &msgAuthorized)
			if err != nil {
				return err
			}

			msg := types.NewMsgRevokeAuthorization(granter, grantee, msgAuthorized)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
	return cmd
}

func GetCmdSendAs(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-as [granter] [msg_tx_json] --from [grantee]",
		Short: "execute tx on behalf of granter account",
		Long:  "execute tx on behalf of granter account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			grantee := cliCtx.FromAddress

			// TODO cleanup this
			//granter, err := sdk.AccAddressFromBech32(args[0])
			//if err != nil {
			//	return err
			//}

			// TODO interactive should look good, consider second arg as optional?
			//generatedTx, err := input.GetString("Enter generated tx json string:", inBuf)

			var stdTx auth.StdTx

			err := cdc.UnmarshalJSON([]byte(args[1]), &stdTx)
			if err != nil {
				return err
			}

			msg := types.NewMsgExecDelegated(grantee, stdTx.Msgs)
			// TODO include the granter as delegated signer in the encoded JSON

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
