package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
)

func GetTipsToFeeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tips-to-fee <aux_signed_tx.json>",
		Short: "tips to fee broadcast the aux signed tx, and sends the tip amount to the broadcaster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			auxTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return err
			}

			tipTx, ok := auxTx.(tx.TipTx)
			if !ok {
				return fmt.Errorf("transcation not supported")
			}

			tip := tipTx.GetTip()
			f := clienttx.NewFactoryCLI(clientCtx, cmd.Flags())
			f.WithTips(sdk.NewCoins(tip.Amount...).String(), tip.Tipper)

			return clienttx.GenerateOrBroadcastTxWithFactory(clientCtx, f, tipTx.GetMsgs()...)
		},
	}

	return cmd
}
