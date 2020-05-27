package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto/multisig"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func GetValidateSignaturesCommand(codec *codec.Codec) *cobra.Command {
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
		RunE:   makeValidateSignaturesCmd(codec),
		Args:   cobra.ExactArgs(1),
	}

	return flags.PostCommands(cmd)[0]
}

func makeValidateSignaturesCmd(cdc *codec.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr, stdTx, err := readStdTxAndInitContexts(cdc, cmd, args[0])
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
	cmd *cobra.Command, cliCtx context.CLIContext, chainID string, stdTx types.StdTx, offline bool,
) bool {
	cmd.Println("Signers:")
	signers := stdTx.GetSigners()

	for i, signer := range signers {
		cmd.Printf("  %v: %v\n", i, signer.String())
	}

	success := true
	sigs := stdTx.Signatures
	cmd.Println("")
	cmd.Println("Signatures:")

	if len(sigs) != len(signers) {
		success = false
	}

	for i, sig := range sigs {
		var (
			multiSigHeader string
			multiSigMsg    string
			sigAddr        = sdk.AccAddress(sig.GetPubKey().Address())
			sigSanity      = "OK"
		)

		if i >= len(signers) || !sigAddr.Equals(signers[i]) {
			sigSanity = "ERROR: signature does not match its respective signer"
			success = false
		}

		// Validate the actual signature over the transaction bytes since we can
		// reach out to a full node to query accounts.
		if !offline && success {
			acc, err := types.NewAccountRetriever(client.Codec).GetAccount(cliCtx, sigAddr)
			if err != nil {
				cmd.Printf("failed to get account: %s\n", sigAddr)
				return false
			}

			sigBytes := types.StdSignBytes(
				chainID, acc.GetAccountNumber(), acc.GetSequence(),
				stdTx.Fee, stdTx.GetMsgs(), stdTx.GetMemo(),
			)

			if ok := sig.GetPubKey().VerifyBytes(sigBytes, sig.Signature); !ok {
				sigSanity = "ERROR: signature invalid"
				success = false
			}
		}

		multiPK, ok := sig.GetPubKey().(multisig.PubKeyMultisigThreshold)
		if ok {
			var multiSig multisig.Multisignature
			cliCtx.Codec.MustUnmarshalBinaryBare(sig.Signature, &multiSig)

			var b strings.Builder
			b.WriteString("\n  MultiSig Signatures:\n")

			for i := 0; i < multiSig.BitArray.Size(); i++ {
				if multiSig.BitArray.GetIndex(i) {
					addr := sdk.AccAddress(multiPK.PubKeys[i].Address().Bytes())
					b.WriteString(fmt.Sprintf("    %d: %s (weight: %d)\n", i, addr, 1))
				}
			}

			multiSigHeader = fmt.Sprintf(" [multisig threshold: %d/%d]", multiPK.K, len(multiPK.PubKeys))
			multiSigMsg = b.String()
		}

		cmd.Printf("  %d: %s\t\t\t[%s]%s%s\n", i, sigAddr.String(), sigSanity, multiSigHeader, multiSigMsg)
	}

	cmd.Println("")

	return success
}

func readStdTxAndInitContexts(cdc *codec.Codec, cmd *cobra.Command, filename string) (
	context.CLIContext, types.TxBuilder, types.StdTx, error,
) {
	stdTx, err := client.ReadStdTxFromFile(cdc, filename)
	if err != nil {
		return context.CLIContext{}, types.TxBuilder{}, types.StdTx{}, err
	}

	inBuf := bufio.NewReader(cmd.InOrStdin())
	cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
	txBldr := types.NewTxBuilderFromCLI(inBuf)

	return cliCtx, txBldr, stdTx, nil
}
