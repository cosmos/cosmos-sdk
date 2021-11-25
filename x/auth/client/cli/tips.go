package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
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

			auxSignerData := &tx.AuxSignerData{}
			err = authclient.ReadAuxSignerData(clientCtx.Codec, auxSignerData, args[0])
			if err != nil {
				return err
			}

			f := clienttx.NewFactoryCLI(clientCtx, cmd.Flags())

			txBuilder := clientCtx.TxConfig.NewTxBuilder()
			err = txBuilder.AddAuxSignerData(*auxSignerData)
			if err != nil {
				return err
			}

			txBuilder.SetFeePayer(clientCtx.FromAddress)
			txBuilder.SetFeeAmount(f.Fees())
			txBuilder.SetGasLimit(f.Gas())

			err = authclient.SignTx(f, clientCtx, clientCtx.FromName, txBuilder, clientCtx.Offline, false)
			if err != nil {
				return err
			}

			txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
			if err != nil {
				return err
			}

			// broadcast to a Tendermint node
			res, err := clientCtx.BroadcastTx(txBytes)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
