package tx

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/spf13/pflag"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// GenerateOrBroadcastTxCLI will either generate and print an unsigned transaction
// or sign it and broadcast it returning an error upon failure.
// TODO: remove the client.Context
func GenerateOrBroadcastTxCLI(ctx client.Context, flagSet *pflag.FlagSet, msgs ...transaction.Msg) error {
	k, err := keyring.NewAutoCLIKeyring(ctx.Keyring)
	if err != nil {
		return err
	}
	// TODO: fulfill with flagSet
	params := TxParameters{
		timeoutHeight:    0,
		chainID:          "",
		memo:             "",
		signMode:         0,
		AccountConfig:    AccountConfig{},
		GasConfig:        GasConfig{},
		FeeConfig:        FeeConfig{},
		ExecutionOptions: ExecutionOptions{},
		ExtensionOptions: ExtensionOptions{},
	}
	txf, err := NewFactory(k, ctx.AddressCodec, ctx, params)
	if err != nil {
		return err
	}

	return GenerateOrBroadcastTxWithFactory(ctx, txf, msgs...)
}

// GenerateOrBroadcastTxWithFactory will either generate and print an unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxWithFactory(clientCtx client.Context, txf Factory, msgs ...transaction.Msg) error {
	// Validate all msgs before generating or broadcasting the tx.
	// We were calling ValidateBasic separately in each CLI handler before.
	// Right now, we're factorizing that call inside this function.
	// ref: https://github.com/cosmos/cosmos-sdk/pull/9236#discussion_r623803504
	for _, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return err
		}
	}

	// If the --aux flag is set, we simply generate and print the AuxSignerData.
	if clientCtx.IsAux {
		auxSignerData, err := makeAuxSignerData(clientCtx, txf, msgs...)
		if err != nil {
			return err
		}

		return clientCtx.PrintString(auxSignerData.String()) // TODO: check this
	}

	if clientCtx.GenerateOnly {
		uTx, err := txf.PrintUnsignedTx(msgs...)
		if err != nil {
			return err
		}
		return clientCtx.PrintString(uTx)
	}

	return BroadcastTx(clientCtx, txf, msgs...)
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func BroadcastTx(clientCtx client.Context, txf Factory, msgs ...sdk.Msg) error {
	txf, err := txf.Prepare()
	if err != nil {
		return err
	}

	if txf.SimulateAndExecute() || clientCtx.Simulate {
		if clientCtx.Offline {
			return errors.New("cannot estimate gas in offline mode")
		}

		_, adjusted, err := CalculateGas(clientCtx, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: txf.Gas()})
	}

	if clientCtx.Simulate {
		return nil
	}

	builder, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return err
	}

	if !clientCtx.SkipConfirm {
		encoder := txf.txConfig.TxJSONEncoder()
		if encoder == nil {
			return errors.New("failed to encode transaction: tx json encoder is nil")
		}

		unsigTx, err := builder.GetTx()
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

	if err = txf.Sign(clientCtx.CmdContext, clientCtx.FromName, builder, true); err != nil {
		return err
	}

	signedTx, err := builder.GetTx()
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

// CalculateGas simulates the execution of a transaction and returns the
// simulation response obtained by the query and the adjusted gas amount.
func CalculateGas(
	clientCtx gogogrpc.ClientConn, txf Factory, msgs ...sdk.Msg,
) (*tx.SimulateResponse, uint64, error) {
	txBytes, err := txf.BuildSimTx(msgs...)
	if err != nil {
		return nil, 0, err
	}

	txSvcClient := tx.NewServiceClient(clientCtx)
	simRes, err := txSvcClient.Simulate(context.Background(), &tx.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, 0, err
	}

	return simRes, uint64(txf.GasAdjustment() * float64(simRes.GasInfo.GasUsed)), nil
}

// makeAuxSignerData generates an AuxSignerData from the client inputs.
func makeAuxSignerData(clientCtx client.Context, f Factory, msgs ...sdk.Msg) (*apitx.AuxSignerData, error) {
	b := NewAuxTxBuilder()
	fromAddress, name, _, err := client.GetFromFields(clientCtx, clientCtx.Keyring, clientCtx.From)
	if err != nil {
		return nil, err
	}

	b.SetAddress(fromAddress.String())
	if f.txParams.offline {
		b.SetAccountNumber(f.AccountNumber())
		b.SetSequence(f.Sequence())
	} else {
		accNum, seq, err := f.accountRetriever.GetAccountNumberSequence(fromAddress)
		if err != nil {
			return nil, err
		}
		b.SetAccountNumber(accNum)
		b.SetSequence(seq)
	}

	err = b.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	err = b.SetSignMode(f.SignMode())
	if err != nil {
		return nil, err
	}

	pubKey, err := f.keybase.GetPubKey(name)
	if err != nil {
		return nil, err
	}

	err = b.SetPubKey(pubKey)
	if err != nil {
		return nil, err
	}

	b.SetChainID(clientCtx.ChainID)
	signBz, err := b.GetSignBytes()
	if err != nil {
		return nil, err
	}

	sig, err := f.keybase.Sign(name, signBz, f.SignMode())
	if err != nil {
		return nil, err
	}
	b.SetSignature(sig)

	return b.GetAuxSignerData()
}

// checkMultipleSigners checks that there can be maximum one DIRECT signer in
// a tx.
func checkMultipleSigners(tx TxWrapper) error {
	directSigners := 0
	sigsV2, err := tx.GetSignatures()
	if err != nil {
		return err
	}
	for _, sig := range sigsV2 {
		directSigners += countDirectSigners(sig.Data)
		if directSigners > 1 {
			return sdkerrors.ErrNotSupported.Wrap("txs signed with CLI can have maximum 1 DIRECT signer")
		}
	}

	return nil
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
