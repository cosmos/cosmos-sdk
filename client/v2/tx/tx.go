package tx

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/pflag"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/client/v2/broadcast"
	"cosmossdk.io/client/v2/broadcast/comet"
	clientcontext "cosmossdk.io/client/v2/context"
	"cosmossdk.io/client/v2/internal/account"
	"cosmossdk.io/client/v2/internal/flags"
	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/codec"
)

// GenerateAndBroadcastTxCLIWithBroadcaster will either generate and print an unsigned transaction
// or sign it and broadcast it with the specified broadcaster returning an error upon failure.
func GenerateAndBroadcastTxCLIWithBroadcaster(
	ctx context.Context,
	conn grpc.ClientConn,
	broadcaster broadcast.Broadcaster,
	msgs ...transaction.Msg,
) ([]byte, error) {
	txf, err := initFactory(ctx, conn, msgs...)
	if err != nil {
		return nil, err
	}

	err = generateTx(txf, msgs...)
	if err != nil {
		return nil, err
	}

	return BroadcastTx(ctx, txf, broadcaster)
}

// GenerateAndBroadcastTxCLI will either generate and print an unsigned transaction
// or sign it and broadcast it using default CometBFT broadcaster, returning an error upon failure.
func GenerateAndBroadcastTxCLI(ctx context.Context, conn grpc.ClientConn, msgs ...transaction.Msg) ([]byte, error) {
	cBroadcaster, err := cometBroadcaster(ctx)
	if err != nil {
		return nil, err
	}

	return GenerateAndBroadcastTxCLIWithBroadcaster(ctx, conn, cBroadcaster, msgs...)
}

// GenerateAndBroadcastTxCLIWithPrompt generates, signs and broadcasts a transaction after prompting the user for confirmation.
// It takes a context, gRPC client connection, prompt function for user confirmation, and transaction messages.
// The prompt function receives the unsigned transaction bytes and returns a boolean indicating user confirmation and any error.
// Returns the broadcast response bytes and any error encountered.
func GenerateAndBroadcastTxCLIWithPrompt(
	ctx context.Context,
	conn grpc.ClientConn,
	prompt func([]byte) (bool, error),
	msgs ...transaction.Msg,
) ([]byte, error) {
	txf, err := initFactory(ctx, conn, msgs...)
	if err != nil {
		return nil, err
	}

	err = generateTx(txf, msgs...)
	if err != nil {
		return nil, err
	}

	confirmed, err := askConfirmation(txf, prompt)
	if err != nil {
		return nil, err
	}
	if !confirmed {
		return nil, nil
	}

	cBroadcaster, err := cometBroadcaster(ctx)
	if err != nil {
		return nil, err
	}

	return BroadcastTx(ctx, txf, cBroadcaster)
}

// GenerateOnly generates an unsigned transaction without broadcasting it.
// It initializes a transaction factory using the provided context, connection and messages,
// then generates an unsigned transaction.
// Returns the unsigned transaction bytes and any error encountered.
func GenerateOnly(ctx context.Context, conn grpc.ClientConn, msgs ...transaction.Msg) ([]byte, error) {
	txf, err := initFactory(ctx, conn)
	if err != nil {
		return nil, err
	}

	return generateOnly(txf, msgs...)
}

// DryRun simulates a transaction without broadcasting it to the network.
// It initializes a transaction factory using the provided context, connection and messages,
// then performs a dry run simulation of the transaction.
// Returns the simulation response bytes and any error encountered.
func DryRun(ctx context.Context, conn grpc.ClientConn, msgs ...transaction.Msg) ([]byte, error) {
	txf, err := initFactory(ctx, conn, msgs...)
	if err != nil {
		return nil, err
	}

	return dryRun(txf, msgs...)
}

// initFactory initializes a new transaction Factory and validates the provided messages.
// It retrieves the client v2 context from the provided context, validates all messages,
// and creates a new transaction Factory using the client context and connection.
// Returns the initialized Factory and any error encountered.
func initFactory(ctx context.Context, conn grpc.ClientConn, msgs ...transaction.Msg) (Factory, error) {
	clientCtx, err := clientcontext.ClientContextFromGoContext(ctx)
	if err != nil {
		return Factory{}, err
	}

	if err := validateMessages(msgs...); err != nil {
		return Factory{}, err
	}

	txf, err := newFactory(*clientCtx, conn)
	if err != nil {
		return Factory{}, err
	}

	return txf, nil
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
		EnabledSignModes:      ctx.EnabledSignModes,
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

	return []byte(fmt.Sprintf(`{"gas_estimate": %d}`, gas)), nil
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

// generateTx generates an unsigned transaction using the provided transaction factory and messages.
// If simulation and execution are enabled, it first calculates the gas requirements.
// It then builds the unsigned transaction with the provided messages.
func generateTx(txf Factory, msgs ...transaction.Msg) error {
	if txf.simulateAndExecute() {
		err := txf.calculateGas(msgs...)
		if err != nil {
			return err
		}
	}

	return txf.BuildUnsignedTx(msgs...)
}

// BroadcastTx attempts to sign and broadcast a transaction using the provided factory and broadcaster.
// GenerateTx must be called first to prepare the transaction for signing.
// This function then signs the transaction using the factory's signing capabilities, encodes it,
// and finally broadcasts it using the provided broadcaster.
func BroadcastTx(ctx context.Context, txf Factory, broadcaster broadcast.Broadcaster) ([]byte, error) {
	if len(txf.tx.msgs) == 0 {
		return nil, errors.New("no messages to broadcast")
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

// cometBroadcaster returns a broadcast.Broadcaster implementation that uses the CometBFT RPC client.
// It extracts the client context from the provided context and uses it to create a CometBFT broadcaster.
func cometBroadcaster(ctx context.Context) (broadcast.Broadcaster, error) {
	c, err := clientcontext.ClientContextFromGoContext(ctx)
	if err != nil {
		return nil, err
	}

	return getCometBroadcaster(c.Cdc, c.Flags)
}

// askConfirmation encodes the transaction as JSON and prompts the user for confirmation using the provided prompter function.
// It returns the user's confirmation response and any error that occurred during the process.
func askConfirmation(txf Factory, prompter func([]byte) (bool, error)) (bool, error) {
	encoder := txf.txConfig.TxJSONEncoder()
	if encoder == nil {
		return false, errors.New("failed to encode transaction: tx json encoder is nil")
	}

	tx, err := txf.getTx()
	if err != nil {
		return false, err
	}

	txBytes, err := encoder(tx)
	if err != nil {
		return false, fmt.Errorf("failed to encode transaction: %w", err)
	}

	return prompter(txBytes)
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
