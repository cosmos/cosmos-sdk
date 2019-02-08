package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetCmdCreateValidator implements the create validator command handler.
// TODO: Add full description
func GetCmdCreateValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			txBldr, msg, err := BuildCreateValidatorMsg(cliCtx, txBldr)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, true)
		},
	}

	cmd.Flags().AddFlagSet(FsPk)
	cmd.Flags().AddFlagSet(FsAmount)
	cmd.Flags().AddFlagSet(fsDescriptionCreate)
	cmd.Flags().AddFlagSet(FsCommissionCreate)
	cmd.Flags().AddFlagSet(FsMinSelfDelegation)
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
// TODO: add full description
func GetCmdEditValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit an existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			valAddr := cliCtx.GetFromAddress()
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

			var newMinSelfDelegation *sdk.Int

			minSelfDelegationString := viper.GetString(FlagMinSelfDelegation)
			if minSelfDelegationString != "" {
				msb, ok := sdk.NewIntFromString(minSelfDelegationString)
				if !ok {
					return fmt.Errorf(staking.ErrMinSelfDelegationInvalid(staking.DefaultCodespace).Error())
				}
				newMinSelfDelegation = &msb
			}

			msg := staking.NewMsgEditValidator(sdk.ValAddress(valAddr), description, newRate, newMinSelfDelegation)

			// build and sign the transaction, then broadcast to Tendermint
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}

	cmd.Flags().AddFlagSet(fsDescriptionEdit)
	cmd.Flags().AddFlagSet(fsCommissionUpdate)

	return cmd
}

// GetCmdDelegate implements the delegate command.
func GetCmdDelegate(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegate [validator-addr] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "delegate liquid tokens to a validator",
		Long: strings.TrimSpace(`Delegate an amount of liquid coins to a validator from your wallet:

$ gaiacli tx staking delegate cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1000stake --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			amount, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			delAddr := cliCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := staking.NewMsgDelegate(delAddr, valAddr, amount)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

// GetCmdRedelegate the begin redelegation command.
func GetCmdRedelegate(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegate [src-validator-addr] [dst-validator-addr] [amount]",
		Short: "redelegate illiquid tokens from one validator to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(`Redelegate an amount of illiquid staking tokens from one validator to another:

$ gaiacli tx staking redelegate cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 100 --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// var err error

			delAddr := cliCtx.GetFromAddress()
			valSrcAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			valDstAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			// get the shares amount
			sharesAmount, err := getShares(args[2], delAddr, valSrcAddr)
			if err != nil {
				return err
			}

			msg := staking.NewMsgBeginRedelegate(delAddr, valSrcAddr, valDstAddr, sharesAmount)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

// GetCmdUnbond implements the unbond validator command.
func GetCmdUnbond(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbond [validator-addr] [amount]",
		Short: "unbond shares from a validator",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(`Unbond an amount of bonded shares from a validator:

$ gaiacli tx staking unbond cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100 --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			delAddr := cliCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// get the shares amount
			sharesAmount, err := getShares(args[1], delAddr, valAddr)
			if err != nil {
				return err
			}

			msg := staking.NewMsgUndelegate(delAddr, valAddr, sharesAmount)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

// BuildCreateValidatorMsg makes a new MsgCreateValidator.
func BuildCreateValidatorMsg(cliCtx context.CLIContext, txBldr authtxb.TxBuilder) (authtxb.TxBuilder, sdk.Msg, error) {
	amounstStr := viper.GetString(FlagAmount)
	amount, err := sdk.ParseCoin(amounstStr)
	if err != nil {
		return txBldr, nil, err
	}

	valAddr := cliCtx.GetFromAddress()
	pkStr := viper.GetString(FlagPubKey)

	pk, err := sdk.GetConsPubKeyBech32(pkStr)
	if err != nil {
		return txBldr, nil, err
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
		return txBldr, nil, err
	}

	// get the initial validator min self delegation
	msbStr := viper.GetString(FlagMinSelfDelegation)
	minSelfDelegation, ok := sdk.NewIntFromString(msbStr)
	if !ok {
		return txBldr, nil, fmt.Errorf(staking.ErrMinSelfDelegationInvalid(staking.DefaultCodespace).Error())
	}

	delAddr := viper.GetString(FlagAddressDelegator)

	var msg sdk.Msg
	if delAddr != "" {
		delAddr, err := sdk.AccAddressFromBech32(delAddr)
		if err != nil {
			return txBldr, nil, err
		}

		msg = staking.NewMsgCreateValidatorOnBehalfOf(
			delAddr, sdk.ValAddress(valAddr), pk, amount, description, commissionMsg, minSelfDelegation,
		)
	} else {
		msg = staking.NewMsgCreateValidator(
			sdk.ValAddress(valAddr), pk, amount, description, commissionMsg, minSelfDelegation,
		)
	}

	if viper.GetBool(client.FlagGenerateOnly) {
		ip := viper.GetString(FlagIP)
		nodeID := viper.GetString(FlagNodeID)
		if nodeID != "" && ip != "" {
			txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
		}
	}
	return txBldr, msg, nil
}
