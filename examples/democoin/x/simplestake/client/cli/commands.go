package cli

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/simplestake"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/crypto/ed25519"
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
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			from, err := cliCtx.GetFromAddress()
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
			var pubKeyEd ed25519.PubKeyEd25519
			copy(pubKeyEd[:], rawPubKey)

			msg := simplestake.NewMsgBond(from, stake, pubKeyEd)

			// Build and sign the transaction, then broadcast to a Tendermint
			// node.
			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
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
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout)

			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			msg := simplestake.NewMsgUnbond(from)

			// Build and sign the transaction, then broadcast to a Tendermint
			// node.
			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}

	return cmd
}
