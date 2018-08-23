package cli

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// create create validator command
func GetCmdCreateValidator(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
			if err != nil {
				return err
			}
			validatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			pkStr := viper.GetString(FlagPubKey)
			if len(pkStr) == 0 {
				return fmt.Errorf("must use --pubkey flag")
			}
			pk, err := sdk.GetValPubKeyBech32(pkStr)
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

			var msg sdk.Msg
			if viper.GetString(FlagAddressDelegator) != "" {
				delegatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
				if err != nil {
					return err
				}
				msg = stake.NewMsgCreateValidatorOnBehalfOf(delegatorAddr, validatorAddr, pk, amount, description)
			} else {
				msg = stake.NewMsgCreateValidator(validatorAddr, pk, amount, description)
			}

			// build and sign the transaction, then broadcast to Tendermint
			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsPk)
	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsDescription)
	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(fsDelegator)
	return cmd
}

// create edit validator command
func GetCmdEditValidator(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit and existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {

			validatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}
			description := stake.Description{
				Moniker:  viper.GetString(FlagMoniker),
				Identity: viper.GetString(FlagIdentity),
				Website:  viper.GetString(FlagWebsite),
				Details:  viper.GetString(FlagDetails),
			}
			msg := stake.NewMsgEditValidator(validatorAddr, description)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsDescription)
	cmd.Flags().AddFlagSet(fsValidator)
	return cmd
}

// delegate command
func GetCmdDelegate(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate",
		Short: "delegate liquid tokens to an validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
			if err != nil {
				return err
			}

			delegatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}
			validatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			msg := stake.NewMsgDelegate(delegatorAddr, validatorAddr, amount)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsValidator)
	return cmd
}

// create edit validator command
func GetCmdRedelegate(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate",
		Short: "redelegate illiquid tokens from one validator to another",
	}
	cmd.AddCommand(
		client.PostCommands(
			GetCmdBeginRedelegate(storeName, cdc),
			GetCmdCompleteRedelegate(cdc),
		)...)
	return cmd
}

// redelegate command
func GetCmdBeginRedelegate(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "begin",
		Short: "begin redelegation",
		RunE: func(cmd *cobra.Command, args []string) error {

			var err error
			delegatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}
			validatorSrcAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidatorSrc))
			if err != nil {
				return err
			}
			validatorDstAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidatorDst))
			if err != nil {
				return err
			}

			// get the shares amount
			sharesAmountStr := viper.GetString(FlagSharesAmount)
			sharesPercentStr := viper.GetString(FlagSharesPercent)
			sharesAmount, err := getShares(storeName, cdc, sharesAmountStr, sharesPercentStr,
				delegatorAddr, validatorSrcAddr)
			if err != nil {
				return err
			}

			msg := stake.NewMsgBeginRedelegate(delegatorAddr, validatorSrcAddr, validatorDstAddr, sharesAmount)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsShares)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsRedelegation)
	return cmd
}

// nolint: gocyclo
// TODO: Make this pass gocyclo linting
func getShares(storeName string, cdc *wire.Codec, sharesAmountStr, sharesPercentStr string,
	delegatorAddr, validatorAddr sdk.AccAddress) (sharesAmount sdk.Rat, err error) {

	switch {
	case sharesAmountStr != "" && sharesPercentStr != "":
		return sharesAmount, errors.Errorf("can either specify the amount OR the percent of the shares, not both")
	case sharesAmountStr == "" && sharesPercentStr == "":
		return sharesAmount, errors.Errorf("can either specify the amount OR the percent of the shares, not both")
	case sharesAmountStr != "":
		sharesAmount, err = sdk.NewRatFromDecimal(sharesAmountStr, types.MaxBondDenominatorPrecision)
		if err != nil {
			return sharesAmount, err
		}
		if !sharesAmount.GT(sdk.ZeroRat()) {
			return sharesAmount, errors.Errorf("shares amount must be positive number (ex. 123, 1.23456789)")
		}
	case sharesPercentStr != "":
		var sharesPercent sdk.Rat
		sharesPercent, err = sdk.NewRatFromDecimal(sharesPercentStr, types.MaxBondDenominatorPrecision)
		if err != nil {
			return sharesAmount, err
		}
		if !sharesPercent.GT(sdk.ZeroRat()) || !sharesPercent.LTE(sdk.OneRat()) {
			return sharesAmount, errors.Errorf("shares percent must be >0 and <=1 (ex. 0.01, 0.75, 1)")
		}

		// make a query to get the existing delegation shares
		key := stake.GetDelegationKey(delegatorAddr, validatorAddr)
		ctx := context.NewCoreContextFromViper()
		resQuery, err := ctx.QueryStore(key, storeName)
		if err != nil {
			return sharesAmount, errors.Errorf("cannot find delegation to determine percent Error: %v", err)
		}
		delegation := types.MustUnmarshalDelegation(cdc, key, resQuery)
		sharesAmount = sharesPercent.Mul(delegation.Shares)
	}
	return
}

// redelegate command
func GetCmdCompleteRedelegate(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete",
		Short: "complete redelegation",
		RunE: func(cmd *cobra.Command, args []string) error {

			delegatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}
			validatorSrcAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidatorSrc))
			if err != nil {
				return err
			}
			validatorDstAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidatorDst))
			if err != nil {
				return err
			}

			msg := stake.NewMsgCompleteRedelegate(delegatorAddr, validatorSrcAddr, validatorDstAddr)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsRedelegation)
	return cmd
}

// create edit validator command
func GetCmdUnbond(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "begin or complete unbonding shares from a validator",
	}
	cmd.AddCommand(
		client.PostCommands(
			GetCmdBeginUnbonding(storeName, cdc),
			GetCmdCompleteUnbonding(cdc),
		)...)
	return cmd
}

// create edit validator command
func GetCmdBeginUnbonding(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "begin",
		Short: "begin unbonding",
		RunE: func(cmd *cobra.Command, args []string) error {

			delegatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}
			validatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			// get the shares amount
			sharesAmountStr := viper.GetString(FlagSharesAmount)
			sharesPercentStr := viper.GetString(FlagSharesPercent)
			sharesAmount, err := getShares(storeName, cdc, sharesAmountStr, sharesPercentStr,
				delegatorAddr, validatorAddr)
			if err != nil {
				return err
			}

			msg := stake.NewMsgBeginUnbonding(delegatorAddr, validatorAddr, sharesAmount)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsShares)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsValidator)
	return cmd
}

// create edit validator command
func GetCmdCompleteUnbonding(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete",
		Short: "complete unbonding",
		RunE: func(cmd *cobra.Command, args []string) error {

			delegatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}
			validatorAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			msg := stake.NewMsgCompleteUnbonding(delegatorAddr, validatorAddr)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsValidator)
	return cmd
}
