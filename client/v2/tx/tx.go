package tx

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/pflag"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/client/v2/internal/account"
	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// GenerateOrBroadcastTxCLI will either generate and print an unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxCLI(ctx client.Context, flagSet *pflag.FlagSet, msgs ...transaction.Msg) error {
	if err := validateMessages(msgs...); err != nil {
		return err
	}

	txf, err := newFactory(ctx, flagSet)
	if err != nil {
		return err
	}

	genOnly, _ := flagSet.GetBool(flagGenerateOnly)
	if genOnly {
		return generateOnly(ctx, txf, msgs...)
	}

	isDryRun, _ := flagSet.GetBool(flagDryRun)
	if isDryRun {
		return dryRun(txf, msgs...)
	}

	return BroadcastTx(ctx, txf, msgs...)
}

// newFactory creates a new transaction Factory based on the provided context and flag set.
// It initializes a new CLI keyring, extracts transaction parameters from the flag set,
// configures transaction settings, and sets up an account retriever for the transaction Factory.
func newFactory(ctx client.Context, flagSet *pflag.FlagSet) (Factory, error) {
	k, err := keyring.NewAutoCLIKeyring(ctx.Keyring, ctx.AddressCodec)
	if err != nil {
		return Factory{}, err
	}

	txConfig, err := NewTxConfig(ConfigOptions{
		AddressCodec:          ctx.AddressCodec,
		Cdc:                   ctx.Codec,
		ValidatorAddressCodec: ctx.ValidatorAddressCodec,
		EnablesSignModes:      ctx.TxConfig.SignModeHandler().SupportedModes(),
	})
	if err != nil {
		return Factory{}, err
	}

	accRetriever := account.NewAccountRetriever(ctx.AddressCodec, ctx, ctx.InterfaceRegistry)

	txf, err := NewFactoryFromFlagSet(flagSet, k, ctx.Codec, accRetriever, txConfig, ctx.AddressCodec, ctx)
	if err != nil {
		return Factory{}, err
	}

	return txf, nil
}

// validateMessages validates all msgs before generating or broadcasting the tx.
// We were calling ValidateBasic separately in each CLI handler before.
// Right now, we're factorizing that call inside this function.
// ref: https://github.com/cosmos/cosmos-sdk/pull/9236#discussion_r623803504
func validateMessages(msgs ...transaction.Msg) error {
	for _, msg := range msgs {
		m, ok := msg.(HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// generateOnly prepares the transaction and prints the unsigned transaction string.
// It first calls Prepare on the transaction factory to set up any necessary pre-conditions.
// If preparation is successful, it generates an unsigned transaction string using the provided messages.
func generateOnly(ctx client.Context, txf Factory, msgs ...transaction.Msg) error {
	uTx, err := txf.UnsignedTxString(msgs...)
	if err != nil {
		return err
	}

	return ctx.PrintString(uTx)
}

// dryRun performs a dry run of the transaction to estimate the gas required.
// It prepares the transaction factory and simulates the transaction with the provided messages.
func dryRun(txf Factory, msgs ...transaction.Msg) error {
	_, gas, err := txf.Simulate(msgs...)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: gas})
	return err
}

// SimulateTx simulates a tx and returns the simulation response obtained by the query.
func SimulateTx(ctx client.Context, flagSet *pflag.FlagSet, msgs ...transaction.Msg) (proto.Message, error) {
	txf, err := newFactory(ctx, flagSet)
	if err != nil {
		return nil, err
	}

	simulation, _, err := txf.Simulate(msgs...)
	return simulation, err
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func BroadcastTx(clientCtx client.Context, txf Factory, msgs ...transaction.Msg) error {
	if txf.simulateAndExecute() {
		err := txf.calculateGas(msgs...)
		if err != nil {
			return err
		}
	}

	err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return err
	}

	if !clientCtx.SkipConfirm {
		encoder := txf.txConfig.TxJSONEncoder()
		if encoder == nil {
			return errors.New("failed to encode transaction: tx json encoder is nil")
		}

		unsigTx, err := txf.getTx()
		if err != nil {
			return err
		}
		txBytes, err := encoder(unsigTx)
		if err != nil {
			return fmt.Errorf("failed to encode transaction: %w", err)
		}

		if err := clientCtx.PrintRaw(txBytes); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\n%s\n", err, txBytes)
		}

		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stderr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v\ncanceled transaction\n", err)
			return err
		}
		if !ok {
			_, _ = fmt.Fprintln(os.Stderr, "canceled transaction")
			return nil
		}
	}

	signedTx, err := txf.sign(clientCtx.CmdContext, true)
	if err != nil {
		return err
	}

	txBytes, err := txf.txConfig.TxEncoder()(signedTx)
	if err != nil {
		return err
	}

	// broadcast to a CometBFT node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}

// countDirectSigners counts the number of DIRECT signers in a signature data.
func countDirectSigners(sigData SignatureData) int {
	switch data := sigData.(type) {
	case *SingleSignatureData:
		if data.SignMode == apitxsigning.SignMode_SIGN_MODE_DIRECT {
			return 1
		}

		return 0
	case *MultiSignatureData:
		directSigners := 0
		for _, d := range data.Signatures {
			directSigners += countDirectSigners(d)
		}

		return directSigners
	default:
		panic("unreachable case")
	}
}

// getSignMode returns the corresponding apitxsigning.SignMode based on the provided mode string.
func getSignMode(mode string) apitxsigning.SignMode {
	switch mode {
	case "direct":
		return apitxsigning.SignMode_SIGN_MODE_DIRECT
	case "direct-aux":
		return apitxsigning.SignMode_SIGN_MODE_DIRECT_AUX
	case "amino-json":
		return apitxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	case "textual":
		return apitxsigning.SignMode_SIGN_MODE_TEXTUAL
	}
	return apitxsigning.SignMode_SIGN_MODE_UNSPECIFIED
}
