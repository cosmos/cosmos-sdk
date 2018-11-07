package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetCmdCreateValidator implements the create validator command handler.
func GetCmdCreateValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			amounstStr := viper.GetString(FlagAmount)
			if amounstStr == "" {
				return fmt.Errorf("Must specify amount to stake using --amount")
			}
			amount, err := sdk.ParseCoin(amounstStr)
			if err != nil {
				return err
			}

			valAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			pkStr := viper.GetString(FlagPubKey)
			if len(pkStr) == 0 {
				return fmt.Errorf("must use --pubkey flag")
			}

			pk, err := sdk.GetConsPubKeyBech32(pkStr)
			if err != nil {
				return err
			}

			if viper.GetString(FlagMoniker) == "" {
				return fmt.Errorf("please enter a moniker for the validator using --moniker")
			}

			description := stake.Description{
				Moniker:  viper.GetString(FlagMoniker),
				Identity: viper.GetString(FlagIdentity),
				Website:  viper.GetString(FlagWebsite),
				Details:  viper.GetString(FlagDetails),
			}

			// get the initial validator commission parameters
			rateStr := viper.GetString(FlagCommissionRate)
			maxRateStr := viper.GetString(FlagCommissionMaxRate)
			maxChangeRateStr := viper.GetString(FlagCommissionMaxChangeRate)
			commissionMsg, err := buildCommissionMsg(rateStr, maxRateStr, maxChangeRateStr)
			if err != nil {
				return err
			}

			var msg sdk.Msg
			if viper.GetString(FlagAddressDelegator) != "" {
				delAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
				if err != nil {
					return err
				}

				msg = stake.NewMsgCreateValidatorOnBehalfOf(
					delAddr, sdk.ValAddress(valAddr), pk, amount, description, commissionMsg,
				)
			} else {
				msg = stake.NewMsgCreateValidator(
					sdk.ValAddress(valAddr), pk, amount, description, commissionMsg,
				)
			}

			if viper.GetBool(FlagGenesisFormat) {
				ip := viper.GetString(FlagIP)
				nodeID := viper.GetString(FlagNodeID)
				if nodeID != "" && ip != "" {
					txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
				}
			}

			if viper.GetBool(FlagGenesisFormat) || cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg}, true)
			}

			// build and sign the transaction, then broadcast to Tendermint
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(FsPk)
	cmd.Flags().AddFlagSet(FsAmount)
	cmd.Flags().AddFlagSet(fsDescriptionCreate)
	cmd.Flags().AddFlagSet(FsCommissionCreate)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().Bool(FlagGenesisFormat, false, "Export the transaction in gen-tx format; it implies --generate-only")
	cmd.Flags().String(FlagIP, "", fmt.Sprintf("Node's public IP. It takes effect only when used in combination with --%s", FlagGenesisFormat))
	cmd.Flags().String(FlagNodeID, "", "Node's ID")
	cmd.MarkFlagRequired(client.FlagFrom)

	return cmd
}

// GetCmdEditValidator implements the create edit validator command.
func GetCmdEditValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit and existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			valAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			description := stake.Description{
				Moniker:  viper.GetString(FlagMoniker),
				Identity: viper.GetString(FlagIdentity),
				Website:  viper.GetString(FlagWebsite),
				Details:  viper.GetString(FlagDetails),
			}

			var newRate *sdk.Dec

			commissionRate := viper.GetString(FlagCommissionRate)
			if commissionRate != "" {
				rate, err := sdk.NewDecFromStr(commissionRate)
				if err != nil {
					return fmt.Errorf("invalid new commission rate: %v", err)
				}

				newRate = &rate
			}

			msg := stake.NewMsgEditValidator(sdk.ValAddress(valAddr), description, newRate)

			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg}, false)
			}

			// build and sign the transaction, then broadcast to Tendermint
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(fsDescriptionEdit)
	cmd.Flags().AddFlagSet(fsCommissionUpdate)

	return cmd
}

// GetCmdDelegate implements the delegate command.
func GetCmdDelegate(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate",
		Short: "delegate liquid tokens to an validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
			if err != nil {
				return err
			}

			delAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			msg := stake.NewMsgDelegate(delAddr, valAddr, amount)

			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg}, false)
			}
			// build and sign the transaction, then broadcast to Tendermint
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(FsAmount)
	cmd.Flags().AddFlagSet(fsValidator)

	return cmd
}

// GetCmdRedelegate implements the redelegate validator command.
func GetCmdRedelegate(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate",
		Short: "redelegate illiquid tokens from one validator to another",
	}

	cmd.AddCommand(
		client.PostCommands(
			GetCmdBeginRedelegate(storeName, cdc),
		)...)

	return cmd
}

// GetCmdBeginRedelegate the begin redelegation command.
func GetCmdBeginRedelegate(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "begin",
		Short: "begin redelegation",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			var err error

			delAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			valSrcAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidatorSrc))
			if err != nil {
				return err
			}

			valDstAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidatorDst))
			if err != nil {
				return err
			}

			// get the shares amount
			sharesAmountStr := viper.GetString(FlagSharesAmount)
			sharesFractionStr := viper.GetString(FlagSharesFraction)
			sharesAmount, err := getShares(
				storeName, cdc, sharesAmountStr, sharesFractionStr,
				delAddr, valSrcAddr,
			)
			if err != nil {
				return err
			}

			msg := stake.NewMsgBeginRedelegate(delAddr, valSrcAddr, valDstAddr, sharesAmount)

			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg}, false)
			}
			// build and sign the transaction, then broadcast to Tendermint
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(fsShares)
	cmd.Flags().AddFlagSet(fsRedelegation)

	return cmd
}

// GetCmdUnbond implements the unbond validator command.
func GetCmdUnbond(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "unbond shares from a validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			delAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			// get the shares amount
			sharesAmountStr := viper.GetString(FlagSharesAmount)
			sharesFractionStr := viper.GetString(FlagSharesFraction)
			sharesAmount, err := getShares(
				storeName, cdc, sharesAmountStr, sharesFractionStr,
				delAddr, valAddr,
			)
			if err != nil {
				return err
			}

			msg := stake.NewMsgBeginUnbonding(delAddr, valAddr, sharesAmount)

			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg}, false)
			}
			// build and sign the transaction, then broadcast to Tendermint
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(fsShares)
	cmd.Flags().AddFlagSet(fsValidator)

	return cmd
}
