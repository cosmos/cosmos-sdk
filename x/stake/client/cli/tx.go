package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/stake"
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
			validatorAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressValidator))
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
			msg := stake.NewMsgCreateValidator(validatorAddr, pk, amount, description)

			// build and sign the transaction, then broadcast to Tendermint
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsPk)
	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsDescription)
	cmd.Flags().AddFlagSet(fsValidator)
	return cmd
}

// create edit validator command
func GetCmdEditValidator(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit and existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {

			validatorAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressValidator))
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

			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsDescription)
	cmd.Flags().AddFlagSet(fsValidator)
	return cmd
}

// create edit validator command
func GetCmdDelegate(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate",
		Short: "delegate coins to an existing validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
			if err != nil {
				return err
			}

			delegatorAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressDelegator))
			validatorAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			msg := stake.NewMsgDelegate(delegatorAddr, validatorAddr, amount)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsValidator)
	return cmd
}

// create edit validator command
func GetCmdUnbond(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "unbond shares from a validator",
		RunE: func(cmd *cobra.Command, args []string) error {

			// check the shares before broadcasting
			sharesStr := viper.GetString(FlagShares)
			var shares sdk.Rat
			if sharesStr != "MAX" {
				var err error
				shares, err = sdk.NewRatFromDecimal(sharesStr)
				if err != nil {
					return err
				}
				if !shares.GT(sdk.ZeroRat()) {
					return fmt.Errorf("shares must be positive integer or decimal (ex. 123, 1.23456789)")
				}
			}

			delegatorAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressDelegator))
			validatorAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			msg := stake.NewMsgUnbond(delegatorAddr, validatorAddr, sharesStr)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsShares)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsValidator)
	return cmd
}
