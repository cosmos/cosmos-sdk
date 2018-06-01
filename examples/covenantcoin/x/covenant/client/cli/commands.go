package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	covenant "github.com/cosmos/cosmos-sdk/examples/covenantcoin/x/covenant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagSettlers  = "settlers"
	flagReceivers = "receivers"
	flagAmount    = "amount"
	flagCovID     = "covid"
	flagReceiver  = "receiver"
)

func CreateCovenantTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_covenant",
		Short: "Create a new Covenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
			sender, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			settlersString := viper.GetString(flagSettlers)
			settlersString = strings.TrimSpace(settlersString)
			if len(settlersString) == 0 {
				return fmt.Errorf("specify comma separated list of settler addresses with --settlers")
			}
			settlersStrs := strings.Split(settlersString, ",")
			var settlers []sdk.Address
			for _, settler := range settlersStrs {
				settlerBytes, err := hex.DecodeString(settler)
				if err != nil {
					return err
				}
				settlers = append(settlers, sdk.Address(settlerBytes))
			}

			receiversString := viper.GetString(flagReceivers)
			receiversString = strings.TrimSpace(receiversString)
			if len(receiversString) == 0 {
				return fmt.Errorf("specify comma separated list of receiver addresses with --receivers")
			}
			receiverStrs := strings.Split(receiversString, ",")
			var receivers []sdk.Address
			for _, receiver := range receiverStrs {
				receiverBytes, err := hex.DecodeString(receiver)
				if err != nil {
					return err
				}
				receivers = append(receivers, sdk.Address(receiverBytes))
			}

			amountString := viper.GetString(flagAmount)
			if len(amountString) == 0 {
				return fmt.Errorf("specify amount as comma separated list of coins with --amount")
			}

			amount, err := sdk.ParseCoins(amountString)
			if err != nil {
				return err
			}

			msg := covenant.MsgCreateCovenant{sender, settlers, receivers, amount}
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}
			newCovID := new(int64)
			err = cdc.UnmarshalBinary(res.DeliverTx.Data, newCovID)
			if err != nil {
				return err
			}
			fmt.Printf("Covenant created with id: %d\n", *newCovID)
			return nil
		},
	}
	cmd.Flags().String(flagSettlers, "", "List of Settler Addresses")
	cmd.Flags().String(flagReceivers, "", "List of Receiver Addresses")
	cmd.Flags().String(flagAmount, "", "Amount to put into covenant")
	return cmd
}

func SettleCovenantTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settle_covenant",
		Short: "Settle and existing Covenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
			settler, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			receiverString := viper.GetString(flagReceiver)
			if len(receiverString) == 0 {
				return fmt.Errorf("specify receiver address with --receiver")
			}
			receiverBytes, err := hex.DecodeString(receiverString)
			if err != nil {
				return err
			}

			receiver := sdk.Address(receiverBytes)

			if !viper.IsSet(flagCovID) {
				return fmt.Errorf("specify Covenant ID with --covid")
			}
			covID := viper.GetInt64(flagCovID)

			msg := covenant.MsgSettleCovenant{covID, settler, receiver}
			_, err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}
			fmt.Printf("Covenant settled with id: %d\n", covID)
			return nil
		},
	}
	cmd.Flags().String(flagCovID, "", "Covenant ID")
	cmd.Flags().String(flagReceiver, "", "Receiver Address")
	return cmd
}
