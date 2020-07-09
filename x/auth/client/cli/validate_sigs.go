package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/spf13/cobra"

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
	sigTx := tx.(authsigning.SigVerifiableTx)
	signModeHandler := clientCtx.TxGenerator.SignModeHandler()

	cmd.Println("Signers:")
	signers := sigTx.GetSigners()

	for i, signer := range signers {
		cmd.Printf("  %v: %v\n", i, signer.String())
	}

	success := true
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		panic(err)
	}
	cmd.Println("")
	cmd.Println("Signatures:")

	if len(sigs) != len(signers) {
		success = false
	}

	for i, sig := range sigs {
		var (
			pubKey         = sig.PubKey
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

			signingData := authsigning.SignerData{
				ChainID:         chainID,
				AccountNumber:   accNum,
				AccountSequence: accSeq,
			}

			switch data := sig.Data.(type) {
			case *signing.SingleSignatureData:
				sigBytes, err := signModeHandler.GetSignBytes(data.SignMode, signingData, tx)
				if err != nil {
					sigSanity = "ERROR: can't get sign bytes"
					success = false
				}

				if ok := pubKey.VerifyBytes(sigBytes, data.Signature); !ok {
					sigSanity = "ERROR: signature invalid"
					success = false
				}
			case *signing.MultiSignatureData:
				multiPK, ok := pubKey.(multisig.PubKey)
				if ok {
					err = multiPK.VerifyMultisignature(func(mode signing.SignMode) ([]byte, error) {
						return signModeHandler.GetSignBytes(mode, signingData, tx)
					}, data)

					if err != nil {
						sigSanity = fmt.Sprintf("ERROR: signature invalid: %v", err)
						success = false
					}

					var b strings.Builder
					b.WriteString("\n  MultiSig Signatures:\n")

					pks := multiPK.GetPubKeys()
					for i := 0; i < data.BitArray.Count(); i++ {
						if data.BitArray.GetIndex(i) {
							addr := sdk.AccAddress(pks[i].Address().Bytes())
							b.WriteString(fmt.Sprintf("    %d: %s (weight: %d)\n", i, addr, 1))
						}
					}

					multiSigHeader = fmt.Sprintf(" [multisig threshold: %d/%d]", multiPK.GetThreshold(), len(pks))
					multiSigMsg = b.String()
				} else {
					sigSanity = "ERROR: expected multisig pub key"
					success = false
				}
			default:
				sigSanity = "ERROR: unexpected ModeInfo"
				success = false
			}
		}

		cmd.Printf("  %d: %s\t\t\t[%s]%s%s\n", i, sigAddr.String(), sigSanity, multiSigHeader, multiSigMsg)
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
	txFactory := tx.NewFactoryCLI(clientCtx, cmd.Flags())

	return clientCtx, txFactory, stdTx, nil
}

// deprecated
func readStdTxAndInitContexts(clientCtx client.Context, cmd *cobra.Command, filename string) (
	client.Context, types.TxBuilder, sdk.Tx, error,
) {
	stdTx, err := authclient.ReadTxFromFile(clientCtx, filename)
	if err != nil {
		return client.Context{}, types.TxBuilder{}, types.StdTx{}, err
	}

	inBuf := bufio.NewReader(cmd.InOrStdin())
	clientCtx = clientCtx.InitWithInput(inBuf)
	txBldr := types.NewTxBuilderFromCLI(inBuf)

	return clientCtx, txBldr, stdTx, nil
}
