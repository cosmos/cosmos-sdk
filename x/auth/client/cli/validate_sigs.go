package cli

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client"
)

func GetValidateSignaturesCommand(cliCtx context.CLIContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-signatures [file]",
		Short: "Validate transactions signatures",
		Long: `Print the addresses that must sign the transaction, those who have already
signed it, and make sure that signatures are in the correct order.

The command would check whether all required signers have signed the transactions, whether
the signatures were collected in the right order, and if the signature is valid over the
given transaction. If the --offline flag is also set, signature validation over the
transaction will be not be performed as that will require RPC communication with a full node.
`,
		PreRun: preSignCmd,
		RunE:   makeValidateSignaturesCmd(cliCtx),
		Args:   cobra.ExactArgs(1),
	}

	return flags.PostCommands(cmd)[0]
}

func makeValidateSignaturesCmd(cliCtx context.CLIContext) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr, stdTx, err := readTxAndInitContexts(cliCtx, cmd, args[0])
		if err != nil {
			return err
		}

		if !printAndValidateSigs(cmd, cliCtx, txBldr.ChainID(), stdTx, cliCtx.Offline) {
			return fmt.Errorf("signatures validation failed")
		}

		return nil
	}
}

// printAndValidateSigs will validate the signatures of a given transaction over its
// expected signers. In addition, if offline has not been supplied, the signature is
// verified over the transaction sign bytes. Returns false if the validation fails.
func printAndValidateSigs(
	cmd *cobra.Command, cliCtx context.CLIContext, chainID string, tx sdk.Tx, offline bool,
) bool {
	//sigTx := tx.(txtypes.SigTx)
	//signModeHandler := cliCtx.TxGenerator.SignModeHandler()
	//
	//cmd.Println("Signers:")
	//signers := sigTx.GetSigners()
	//
	//for i, signer := range signers {
	//	cmd.Printf("  %v: %v\n", i, signer.String())
	//}
	//
	//success := true
	//sigs := sigTx.GetSignatures()
	//pubKeys := sigTx.GetPubKeys()
	//cmd.Println("")
	//cmd.Println("Signatures:")
	//
	//if len(sigs) != len(signers) {
	//	success = false
	//}
	//
	//for i, sig := range sigs {
	//	var (
	//		pubKey         = pubKeys[i]
	//		multiSigHeader string
	//		multiSigMsg    string
	//		sigAddr        = sdk.AccAddress(pubKey.Address())
	//		sigSanity      = "OK"
	//	)
	//
	//	if i >= len(signers) || !sigAddr.Equals(signers[i]) {
	//		sigSanity = "ERROR: signature does not match its respective signer"
	//		success = false
	//	}
	//
	//	// Validate the actual signature over the transaction bytes since we can
	//	// reach out to a full node to query accounts.
	//	if !offline && success {
	//		accNum, accSeq, err := cliCtx.AccountRetriever.GetAccountNumberSequence(cliCtx, sigAddr)
	//		if err != nil {
	//			cmd.Printf("failed to get account: %s\n", sigAddr)
	//			return false
	//		}
	//
	//		signingData := txtypes.SigningData{
	//			PublicKey:       pubKey,
	//			ChainID:         chainID,
	//			AccountNumber:   accNum,
	//			AccountSequence: accSeq,
	//		}
	//
	//		if protoTx, ok := tx.(txtypes.ProtoTx); ok {
	//			modeInfo := protoTx.GetAuthInfo().SignerInfos[i].ModeInfo
	//			switch modeInfo := modeInfo.Sum.(type) {
	//			case *txtypes.ModeInfo_Single_:
	//				signingData.ModeInfo = modeInfo.Single
	//				sigBytes, err := signModeHandler.GetSignBytes(signingData, tx)
	//				if err != nil {
	//					sigSanity = "ERROR: can't get sign bytes"
	//					success = false
	//				}
	//
	//				if ok := pubKey.VerifyBytes(sigBytes, sig); !ok {
	//					sigSanity = "ERROR: signature invalid"
	//					success = false
	//				}
	//			case *txtypes.ModeInfo_Multi_:
	//				multisigs, err := types.DecodeMultisignatures(sig)
	//				if err != nil {
	//					sigSanity = "ERROR: can't decoded multisignature"
	//					success = false
	//				}
	//
	//				if multiPK, ok := pubKey.(multisig2.MultisigPubKey); ok {
	//					multiPK.VerifyMultisignature(func(single *txtypes.ModeInfo_Single) ([]byte, error) {
	//						signingData.ModeInfo = single
	//						return signModeHandler.GetSignBytes(signingData, tx)
	//					}, multisig2.DecodedMultisignature{
	//						ModeInfo:   modeInfo.Multi,
	//						Signatures: multisigs,
	//					})
	//				}
	//			default:
	//				sigSanity = "ERROR: unexpected ModeInfo"
	//				success = false
	//			}
	//		} else {
	//			sigBytes, err := signModeHandler.GetSignBytes(signingData, tx)
	//			if err != nil {
	//				sigSanity = "ERROR: can't get sign bytes"
	//				success = false
	//			}
	//
	//			if ok := pubKey.VerifyBytes(sigBytes, sig); !ok {
	//				sigSanity = "ERROR: signature invalid"
	//				success = false
	//			}
	//		}
	//	}
	//
	//	multiPK, ok := pubKey.(multisig.PubKeyMultisigThreshold)
	//	if ok {
	//		var multiSig multisig.Multisignature
	//
	//		panic("TODO: update this:")
	//		cliCtx.Codec.MustUnmarshalBinaryBare(sig, &multiSig)
	//
	//		var b strings.Builder
	//		b.WriteString("\n  MultiSig Signatures:\n")
	//
	//		for i := 0; i < multiSig.BitArray.Size(); i++ {
	//			if multiSig.BitArray.GetIndex(i) {
	//				addr := sdk.AccAddress(multiPK.PubKeys[i].Address().Bytes())
	//				b.WriteString(fmt.Sprintf("    %d: %s (weight: %d)\n", i, addr, 1))
	//			}
	//		}
	//
	//		multiSigHeader = fmt.Sprintf(" [multisig threshold: %d/%d]", multiPK.K, len(multiPK.PubKeys))
	//		multiSigMsg = b.String()
	//	}
	//
	//	cmd.Printf("  %d: %s\t\t\t[%s]%s%s\n", i, sigAddr.String(), sigSanity, multiSigHeader, multiSigMsg)
	//}
	//
	//cmd.Println("")
	//
	//return success
	panic("TODO")
}

func readTxAndInitContexts(cliCtx context.CLIContext, cmd *cobra.Command, filename string) (context.CLIContext, tx.Factory, sdk.Tx, error) {
	stdTx, err := client.ReadTxFromFile(cliCtx, filename)
	if err != nil {
		return cliCtx, tx.Factory{}, nil, err
	}

	inBuf := bufio.NewReader(cmd.InOrStdin())
	cliCtx = cliCtx.InitWithInput(inBuf)
	txFactory := tx.NewFactoryFromCLI(inBuf)

	return cliCtx, txFactory, stdTx, nil
}
