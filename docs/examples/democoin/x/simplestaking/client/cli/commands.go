package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/docs/examples/democoin/x/simplestaking"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/crypto/ed25519"
)

const (
	flagStake     = "staking"
	flagValidator = "validator"
)

// simple bond tx
func BondTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond",
		Short: "Bond to a validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			stakingString := viper.GetString(flagStake)
			if len(stakingString) == 0 {
				return fmt.Errorf("specify coins to bond with --staking")
			}

			valString := viper.GetString(flagValidator)
			if len(valString) == 0 {
				return fmt.Errorf("specify pubkey to bond to with --validator")
			}

			staking, err := sdk.ParseCoin(stakingString)
			if err != nil {
				return err
			}

			// TODO: bech32 ...
			rawPubKey, err := hex.DecodeString(valString)
			if err != nil {
				return err
			}
			var pubKeyEd ed25519.PubKeyEd25519
			copy(pubKeyEd[:], rawPubKey)

			return cliCtx.MessageOutput(simplestaking.NewMsgBond(cliCtx.FromAddr(), staking, pubKeyEd))
		},
	}

	cmd.Flags().String(flagStake, "", "Amount of coins to stake")
	cmd.Flags().String(flagValidator, "", "Validator address to stake")

	return cmd
}

// simple unbond tx
func UnbondTxCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbond",
		Short: "Unbond from a validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)
			return cliCtx.MessageOutput(simplestaking.NewMsgUnbond(cliCtx.FromAddr()))
		},
	}
}
