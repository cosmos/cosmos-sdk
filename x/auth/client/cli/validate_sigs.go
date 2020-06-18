package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	multisig2 "github.com/cosmos/cosmos-sdk/crypto/multisig"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
)

func GetValidateSignaturesCommand(clientCtx client.Context) *cobra.Command {
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
		RunE:   makeValidateSignaturesCmd(clientCtx),
		Args:   cobra.ExactArgs(1),
	}

	return flags.PostCommands(cmd)[0]
}

func makeValidateSignaturesCmd(clientCtx client.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, txBldr, stdTx, err := readTxAndInitContexts(clientCtx, cmd, args[0])
		if err != nil {
			return err
		}

		if !printAndValidateSigs(cmd, clientCtx, txBldr.ChainID(), stdTx, clientCtx.Offline) {
			return fmt.Errorf("signatures validation failed")
		}

		return nil
	}
}

// printAndValidateSigs will validate the signatures of a given transaction over its
// expected signers. In addition, if offline has not been supplied, the signature is
// verified over the transaction sign bytes. Returns false if the validation fails.
func printAndValidateSigs(
	cmd *cobra.Command, clientCtx client.Context, chainID string, tx sdk.Tx, offline bool,
) bool {
	sigTx := tx.(txtypes.SigTx)
	signModeHandler := clientCtx.TxGenerator.SignModeHandler()

	cmd.Println("Signers:")
	signers := sigTx.GetSigners()

	for i, signer := range signers {
		cmd.Printf("  %v: %v\n", i, signer.String())
	}

	success := true
	sigs, err := sigTx.GetSignatureData()
	if err != nil {
		panic(err)
	}
	pubKeys := sigTx.GetPubKeys()
	cmd.Println("")
	cmd.Println("Signatures:")

	if len(sigs) != len(signers) {
		success = false
	}

	for i, sig := range sigs {
		var (
			pubKey         = pubKeys[i]
			multiSigHeader string
			multiSigMsg    string
			sigAddr        = sdk.AccAddress(pubKey.Address())
			sigSanity      = "OK"
		)

		if i >= len(signers) || !sigAddr.Equals(signers[i]) {
			sigSanity = "ERROR: signature does not match its respective signer"
			success = false
		}

		// Validate the actual signature over the transaction bytes since we can
		// reach out to a full node to query accounts.
		if !offline && success {
			accNum, accSeq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, sigAddr)
			if err != nil {
				cmd.Printf("failed to get account: %s\n", sigAddr)
				return false
			}

			signingData := txtypes.SigningData{
				PublicKey:       pubKey,
				ChainID:         chainID,
				AccountNumber:   accNum,
				AccountSequence: accSeq,
			}

			switch sig := sig.(type) {
			case *txtypes.SingleSignatureData:
				sigBytes, err := signModeHandler.GetSignBytes(sig.SignMode, signingData, tx)
				if err != nil {
					sigSanity = "ERROR: can't get sign bytes"
					success = false
				}

				if ok := pubKey.VerifyBytes(sigBytes, sig.Signature); !ok {
					sigSanity = "ERROR: signature invalid"
					success = false
				}
			case *txtypes.MultiSignatureData:
				multiPK, ok := pubKey.(multisig2.MultisigPubKey)
				if ok {
					if !multiPK.VerifyMultisignature(func(mode txtypes.SignMode) ([]byte, error) {
						signingData.Mode = mode
						return signModeHandler.GetSignBytes(mode, signingData, tx)
					}, sig) {
						sigSanity = "ERROR: signature invalid"
						success = false
					}

					var b strings.Builder
					b.WriteString("\n  MultiSig Signatures:\n")

					pks := multiPK.GetPubKeys()
					for i := 0; i < sig.BitArray.Size(); i++ {
						if sig.BitArray.GetIndex(i) {
							addr := sdk.AccAddress(pks[i].Address().Bytes())
							b.WriteString(fmt.Sprintf("    %d: %s (weight: %d)\n", i, addr, 1))
						}
					}

					multiSigHeader = fmt.Sprintf(" [multisig threshold: %d/%d]", multiPK.Threshold(), len(pks))
					multiSigMsg = b.String()
					cmd.Printf("  %d: %s\t\t\t[%s]%s%s\n", i, sigAddr.String(), sigSanity, multiSigHeader, multiSigMsg)
				} else {
					sigSanity = "ERROR: expected multisig pub key"
					success = false
				}
			default:
				sigSanity = "ERROR: unexpected ModeInfo"
				success = false
			}
		}
	}

	cmd.Println("")

	return success
}

func readTxAndInitContexts(clientCtx client.Context, cmd *cobra.Command, filename string) (client.Context, tx.Factory, sdk.Tx, error) {
	stdTx, err := authclient.ReadTxFromFile(clientCtx, filename)
	if err != nil {
		return clientCtx, tx.Factory{}, nil, err
	}

	inBuf := bufio.NewReader(cmd.InOrStdin())
	clientCtx = clientCtx.InitWithInput(inBuf)
	txFactory := tx.NewFactoryFromCLI(inBuf)

	return clientCtx, txFactory, stdTx, nil
}
