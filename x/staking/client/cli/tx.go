package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetCmdCreateValidator implements the create validator command handler.
func GetCmdCreateValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			msg, err := BuildCreateValidatorMsg(cliCtx)
			if err != nil {
				return err
			}

			return cliCtx.MessageOutput(msg)
		},
	}

	cmd.Flags().AddFlagSet(FsPk)
	cmd.Flags().AddFlagSet(FsAmount)
	cmd.Flags().AddFlagSet(fsDescriptionCreate)
	cmd.Flags().AddFlagSet(FsCommissionCreate)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().String(FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", client.FlagGenerateOnly))
	cmd.Flags().String(FlagNodeID, "", "The node's ID")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagAmount)
	cmd.MarkFlagRequired(FlagPubKey)
	cmd.MarkFlagRequired(FlagMoniker)

	return cmd
}

// GetCmdEditValidator implements the create edit validator command.
func GetCmdEditValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit an existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			description := staking.Description{
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

			msg := staking.NewMsgEditValidator(cliCtx.FromValAddr(), description, newRate)
			return cliCtx.MessageOutput(msg)
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
		Short: "delegate liquid tokens to a validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			return cliCtx.MessageOutput(staking.NewMsgDelegate(cliCtx.FromAddr(), valAddr, amount))
		},
	}

	cmd.Flags().AddFlagSet(FsAmount)
	cmd.Flags().AddFlagSet(fsValidator)

	return cmd
}

// GetCmdRedelegate the begin redelegation command.
func GetCmdRedelegate(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate",
		Short: "redelegate illiquid tokens from one validator to another",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			var err error

			delAddr := cliCtx.FromAddr()

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

			msg := staking.NewMsgBeginRedelegate(delAddr, valSrcAddr, valDstAddr, sharesAmount)
			return cliCtx.MessageOutput(msg)
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
			cliCtx := context.NewCLIContextTx(cdc)

			delAddr := cliCtx.FromAddr()

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

			return cliCtx.MessageOutput(staking.NewMsgUndelegate(delAddr, valAddr, sharesAmount))
		},
	}

	cmd.Flags().AddFlagSet(fsShares)
	cmd.Flags().AddFlagSet(fsValidator)

	return cmd
}

// BuildCreateValidatorMsg makes a new MsgCreateValidator.
func BuildCreateValidatorMsg(cliCtx *context.CLIContext) (sdk.Msg, error) {
	amounstStr := viper.GetString(FlagAmount)
	amount, err := sdk.ParseCoin(amounstStr)
	if err != nil {
		return nil, err
	}

	pkStr := viper.GetString(FlagPubKey)
	pk, err := sdk.GetConsPubKeyBech32(pkStr)
	if err != nil {
		return nil, err
	}

	description := staking.NewDescription(
		viper.GetString(FlagMoniker),
		viper.GetString(FlagIdentity),
		viper.GetString(FlagWebsite),
		viper.GetString(FlagDetails),
	)

	// get the initial validator commission parameters
	rateStr := viper.GetString(FlagCommissionRate)
	maxRateStr := viper.GetString(FlagCommissionMaxRate)
	maxChangeRateStr := viper.GetString(FlagCommissionMaxChangeRate)
	commissionMsg, err := buildCommissionMsg(rateStr, maxRateStr, maxChangeRateStr)
	if err != nil {
		return nil, err
	}

	delAddr := viper.GetString(FlagAddressDelegator)

	var msg sdk.Msg
	if delAddr != "" {
		delAddr, err := sdk.AccAddressFromBech32(delAddr)
		if err != nil {
			return nil, err
		}

		msg = staking.NewMsgCreateValidatorOnBehalfOf(
			delAddr, cliCtx.FromValAddr(), pk, amount, description, commissionMsg,
		)
	} else {
		msg = staking.NewMsgCreateValidator(
			cliCtx.FromValAddr(), pk, amount, description, commissionMsg,
		)
	}

	if viper.GetBool(client.FlagGenerateOnly) {
		ip := viper.GetString(FlagIP)
		nodeID := viper.GetString(FlagNodeID)
		if nodeID != "" && ip != "" {
			cliCtx.TxBldr = cliCtx.TxBldr.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
		}
	}
	return msg, nil
}
