package tx

import (
	"context"
	"fmt"
	"github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/pflag"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	clientcontext "cosmossdk.io/client/v2/autocli/context"
	"cosmossdk.io/client/v2/broadcast"
	"cosmossdk.io/client/v2/broadcast/comet"
	"cosmossdk.io/client/v2/internal/account"
	"cosmossdk.io/client/v2/internal/flags"
	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/codec"
)

// GenerateOrBroadcastTxCLIWithBroadcaster will either generate and print an unsigned transaction
// or sign it and broadcast it with the specified broadcaster returning an error upon failure.
func GenerateOrBroadcastTxCLIWithBroadcaster(
	ctx context.Context,
	conn grpc.ClientConn,
	broadcaster broadcast.Broadcaster,
	msgs ...transaction.Msg,
) ([]byte, error) {
	clientCtx, err := clientcontext.ClientContextFromGoContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := validateMessages(msgs...); err != nil {
		return nil, err
	}

	txf, err := newFactory(*clientCtx, conn)
	if err != nil {
		return nil, err
	}

	genOnly, _ := clientCtx.Flags.GetBool(flagGenerateOnly)
	if genOnly {
		return generateOnly(txf, msgs...)
	}

	isDryRun, _ := clientCtx.Flags.GetBool(flagDryRun)
	if isDryRun {
		return dryRun(txf, msgs...)
	}

	return BroadcastTx(ctx, txf, broadcaster, msgs...)
}

// GenerateOrBroadcastTxCLI will either generate and print an unsigned transaction
// or sign it and broadcast it using default CometBFT broadcaster, returning an error upon failure.
func GenerateOrBroadcastTxCLI(
	ctx context.Context,
	conn grpc.ClientConn,
	msgs ...transaction.Msg,
) ([]byte, error) {
	c, err := clientcontext.ClientContextFromGoContext(ctx)
	if err != nil {
		return nil, err
	}

	cometBroadcaster, err := getCometBroadcaster(c.Cdc, c.Flags)
	if err != nil {
		return nil, err
	}

	return GenerateOrBroadcastTxCLIWithBroadcaster(ctx, conn, cometBroadcaster, msgs...)
}

// getCometBroadcaster returns a new CometBFT broadcaster based on the provided context and flag set.
func getCometBroadcaster(cdc codec.Codec, flagSet *pflag.FlagSet) (broadcast.Broadcaster, error) {
	url, _ := flagSet.GetString(flags.FlagNode)
	mode, _ := flagSet.GetString(flags.FlagBroadcastMode)
	return comet.NewCometBFTBroadcaster(url, mode, cdc)
}

// newFactory creates a new transaction Factory based on the provided context and flag set.
// It initializes a new CLI keyring, extracts transaction parameters from the flag set,
// configures transaction settings, and sets up an account retriever for the transaction Factory.
func newFactory(ctx clientcontext.Context, conn grpc.ClientConn) (Factory, error) {
	txConfig, err := NewTxConfig(ConfigOptions{
		AddressCodec:          ctx.AddressCodec,
		Cdc:                   ctx.Cdc,
		ValidatorAddressCodec: ctx.ValidatorAddressCodec,
		EnabledSignModes:      ctx.EnabledSignmodes,
	})
	if err != nil {
		return Factory{}, err
	}

	accRetriever := account.NewAccountRetriever(ctx.AddressCodec, conn, ctx.Cdc.InterfaceRegistry())

	txf, err := NewFactoryFromFlagSet(ctx.Flags, ctx.Keyring, ctx.Cdc, accRetriever, txConfig, ctx.AddressCodec, conn)
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
func generateOnly(txf Factory, msgs ...transaction.Msg) ([]byte, error) {
	uTx, err := txf.UnsignedTxString(msgs...)
	if err != nil {
		return nil, err
	}

	return []byte(uTx), nil
}

// dryRun performs a dry run of the transaction to estimate the gas required.
// It prepares the transaction factory and simulates the transaction with the provided messages.
func dryRun(txf Factory, msgs ...transaction.Msg) ([]byte, error) {
	_, gas, err := txf.Simulate(msgs...)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("%s\n", GasEstimateResponse{GasEstimate: gas})), nil
}

// SimulateTx simulates a tx and returns the simulation response obtained by the query.
func SimulateTx(ctx clientcontext.Context, conn grpc.ClientConn, msgs ...transaction.Msg) (proto.Message, error) {
	txf, err := newFactory(ctx, conn)
	if err != nil {
		return nil, err
	}

	simulation, _, err := txf.Simulate(msgs...)
	return simulation, err
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func BroadcastTx(ctx context.Context, txf Factory, broadcaster broadcast.Broadcaster, msgs ...transaction.Msg) ([]byte, error) {
	if txf.simulateAndExecute() {
		err := txf.calculateGas(msgs...)
		if err != nil {
			return nil, err
		}
	}

	err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	signedTx, err := txf.sign(ctx, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := txf.txConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, err
	}

	return broadcaster.Broadcast(ctx, txBytes)
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
