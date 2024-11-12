package tx

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"os"

	"github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/pflag"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/client/v2/autocli/print"
	"cosmossdk.io/client/v2/broadcast"
	"cosmossdk.io/client/v2/broadcast/comet"
	"cosmossdk.io/client/v2/internal/account"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GenerateOrBroadcastTxCLIWithBroadcaster will either generate and print an unsigned transaction
// or sign it and broadcast it with the specified broadcaster returning an error upon failure.
func GenerateOrBroadcastTxCLIWithBroadcaster(
	flagSet *pflag.FlagSet,
	printer *print.Printer,
	keybase keyring.Keyring,
	cdc codec.Codec,
	addressCodec, validatorCodec address.Codec,
	enablesSignModes []apitxsigning.SignMode,
	conn grpc.ClientConn,
	broadcaster broadcast.Broadcaster,
	msgs ...transaction.Msg,
) error {
	if err := validateMessages(msgs...); err != nil {
		return err
	}

	txf, err := newFactory(keybase, cdc, addressCodec, validatorCodec, enablesSignModes, conn, flagSet)
	if err != nil {
		return err
	}

	genOnly, _ := flagSet.GetBool(flagGenerateOnly)
	if genOnly {
		return generateOnly(printer, txf, msgs...)
	}

	isDryRun, _ := flagSet.GetBool(flagDryRun)
	if isDryRun {
		return dryRun(printer, txf, msgs...)
	}

	skipConfirm, _ := flagSet.GetBool("yes")
	return BroadcastTx(printer, txf, broadcaster, skipConfirm, msgs...)
}

// GenerateOrBroadcastTxCLI will either generate and print an unsigned transaction
// or sign it and broadcast it using default CometBFT broadcaster, returning an error upon failure.
func GenerateOrBroadcastTxCLI(
	flagSet *pflag.FlagSet,
	printer *print.Printer,
	keybase keyring.Keyring,
	cdc codec.Codec,
	addressCodec, validatorCodec address.Codec,
	enablesSignModes []apitxsigning.SignMode,
	conn grpc.ClientConn,
	msgs ...transaction.Msg,
) error {
	cometBroadcaster, err := getCometBroadcaster(cdc, cdc.InterfaceRegistry(), flagSet)
	if err != nil {
		return err
	}

	return GenerateOrBroadcastTxCLIWithBroadcaster(flagSet, printer, keybase, cdc, addressCodec, validatorCodec, enablesSignModes, conn, cometBroadcaster, msgs...)
}

// getCometBroadcaster returns a new CometBFT broadcaster based on the provided context and flag set.
func getCometBroadcaster(cdc codec.Codec, ir types.InterfaceRegistry, flagSet *pflag.FlagSet) (broadcast.Broadcaster, error) {
	url, _ := flagSet.GetString("node")
	mode, _ := flagSet.GetString("broadcast-mode")
	return comet.NewCometBFTBroadcaster(url, mode, cdc, ir)
}

// newFactory creates a new transaction Factory based on the provided context and flag set.
// It initializes a new CLI keyring, extracts transaction parameters from the flag set,
// configures transaction settings, and sets up an account retriever for the transaction Factory.
func newFactory(
	keybase keyring.Keyring,
	cdc codec.Codec,
	addressCodec, validatorCodec address.Codec,
	enablesSignModes []apitxsigning.SignMode,
	conn grpc.ClientConn,
	flagSet *pflag.FlagSet,
) (Factory, error) {
	txConfig, err := NewTxConfig(ConfigOptions{
		AddressCodec:          addressCodec,
		Cdc:                   cdc,
		ValidatorAddressCodec: validatorCodec,
		EnabledSignModes:      enablesSignModes,
	})
	if err != nil {
		return Factory{}, err
	}

	accRetriever := account.NewAccountRetriever(addressCodec, conn, cdc.InterfaceRegistry())

	txf, err := NewFactoryFromFlagSet(flagSet, keybase, cdc, accRetriever, txConfig, addressCodec, conn)
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
func generateOnly(printer *print.Printer, txf Factory, msgs ...transaction.Msg) error {
	uTx, err := txf.UnsignedTxString(msgs...)
	if err != nil {
		return err
	}

	return printer.PrintString(uTx)
}

// dryRun performs a dry run of the transaction to estimate the gas required.
// It prepares the transaction factory and simulates the transaction with the provided messages.
func dryRun(printer *print.Printer, txf Factory, msgs ...transaction.Msg) error {
	_, gas, err := txf.Simulate(msgs...)
	if err != nil {
		return err
	}

	return printer.PrintString(fmt.Sprintf("%s\n", GasEstimateResponse{GasEstimate: gas}))
}

// SimulateTx simulates a tx and returns the simulation response obtained by the query.
func SimulateTx(
	keybase keyring.Keyring,
	cdc codec.Codec,
	addressCodec, validatorCodec address.Codec,
	enablesSignModes []apitxsigning.SignMode,
	conn grpc.ClientConn,
	flagSet *pflag.FlagSet,
	msgs ...transaction.Msg,
) (proto.Message, error) {
	txf, err := newFactory(keybase, cdc, addressCodec, validatorCodec, enablesSignModes, conn, flagSet)
	if err != nil {
		return nil, err
	}

	simulation, _, err := txf.Simulate(msgs...)
	return simulation, err
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func BroadcastTx(printer *print.Printer, txf Factory, broadcaster broadcast.Broadcaster, skipConfirm bool, msgs ...transaction.Msg) error {
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

	if !skipConfirm {
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

		if err := printer.PrintRaw(txBytes); err != nil {
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

	signedTx, err := txf.sign(context.Background(), true) // TODO: pass ctx from upper call
	if err != nil {
		return err
	}

	txBytes, err := txf.txConfig.TxEncoder()(signedTx)
	if err != nil {
		return err
	}

	res, err := broadcaster.Broadcast(context.Background(), txBytes)
	if err != nil {
		return err
	}

	return printer.PrintBytes(res)
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
