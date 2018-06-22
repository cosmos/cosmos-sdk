package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"

	"github.com/cosmos/cosmos-sdk/examples/democoin/x/simplestake"
)

const (
	flagStake     = "stake"
	flagValidator = "validator"
)

// simple bond tx
func BondTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond",
		Short: "Bond to a validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper()

			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			stakeString := viper.GetString(flagStake)
			if len(stakeString) == 0 {
				return fmt.Errorf("specify coins to bond with --stake")
			}

			valString := viper.GetString(flagValidator)
			if len(valString) == 0 {
				return fmt.Errorf("specify pubkey to bond to with --validator")
			}

			stake, err := sdk.ParseCoin(stakeString)
			if err != nil {
				return err
			}

			// TODO: bech32 ...
			rawPubKey, err := hex.DecodeString(valString)
			if err != nil {
				return err
			}
			var pubKeyEd crypto.PubKeyEd25519
			copy(pubKeyEd[:], rawPubKey)

			msg := simplestake.NewMsgBond(from, stake, pubKeyEd)

			return sendMsg(cdc, msg)
		},
	}
	cmd.Flags().String(flagStake, "", "Amount of coins to stake")
	cmd.Flags().String(flagValidator, "", "Validator address to stake")
	return cmd
}

// simple unbond tx
func UnbondTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "Unbond from a validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			from, err := context.NewCoreContextFromViper().GetFromAddress()
			if err != nil {
				return err
			}
			msg := simplestake.NewMsgUnbond(from)
			return sendMsg(cdc, msg)
		},
	}
	return cmd
}

func sendMsg(cdc *wire.Codec, msg sdk.Msg) error {
	ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
	res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
	if err != nil {
		return err
	}

	fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
	return nil
}
