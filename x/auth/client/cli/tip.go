package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
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
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid transaction")
			}

			f := clienttx.NewFactoryCLI(clientCtx, cmd.Flags())

			tipperSigsV2, err := tipTx.(authsigning.SigVerifiableTx).GetSignaturesV2()
			if err != nil {
				return err
			}
			txBuilder := clientCtx.TxConfig.NewTxBuilder()

			txBuilder.SetMsgs(tipTx.GetMsgs()...)
			txBuilder.SetFeePayer(clientCtx.FromAddress)
			txBuilder.SetFeeAmount(f.Fees())
			txBuilder.SetGasLimit(f.Gas())
			txBuilder.SetTip(tipTx.GetTip())
			err = txBuilder.SetSignatures(tipperSigsV2...)
			if err != nil {
				return err
			}

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
