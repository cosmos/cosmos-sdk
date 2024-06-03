package tx

import (
	"bufio"
	"context"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	tx2 "cosmossdk.io/client/v2/internal"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/input"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogogrpc "github.com/cosmos/gogoproto/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/spf13/pflag"
	"os"
)

// GenerateOrBroadcastTxCLI will either generate and print an unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxCLI(clientCtx tx2.Context, flagSet *pflag.FlagSet, msgs ...sdk.Msg) error {
	txf, err := NewFactoryCLI(clientCtx, flagSet)
	if err != nil {
		return err
	}

	return GenerateOrBroadcastTxWithFactory(clientCtx, txf, msgs...)
}

// GenerateOrBroadcastTxWithFactory will either generate and print an unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxWithFactory(clientCtx tx2.Context, txf Factory, msgs ...sdk.Msg) error {
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

		return clientCtx.PrintProto(&auxSignerData)
	}

	if clientCtx.GenerateOnly {
		return txf.PrintUnsignedTx(clientCtx, msgs...)
	}

	return BroadcastTx(clientCtx, txf, msgs...)
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func BroadcastTx(clientCtx tx2.Context, txf Factory, msgs ...sdk.Msg) error {
	txf, err := txf.Prepare(clientCtx)
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

	tx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return err
	}

	if !clientCtx.SkipConfirm {
		encoder := txf.txConfig.TxJSONEncoder()
		if encoder == nil {
			return errors.New("failed to encode transaction: tx json encoder is nil")
		}

		txBytes, err := encoder(tx.GetTx())
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

	if err = txf.Sign(clientCtx.CmdContext, clientCtx.FromName, tx, true); err != nil {
		return err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(tx.GetTx())
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
func makeAuxSignerData(clientCtx tx2.Context, f Factory, msgs ...sdk.Msg) (*apitx.AuxSignerData, error) {
	b := NewAuxTxBuilder()
	fromAddress, name, _, err := tx2.GetFromFields(clientCtx, clientCtx.Keyring, clientCtx.From)
	if err != nil {
		return nil, err
	}

	b.SetAddress(fromAddress.String())
	if f.txParams.offline {
		b.SetAccountNumber(f.AccountNumber())
		b.SetSequence(f.Sequence())
	} else {
		accNum, seq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, fromAddress)
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

	pubKey, err := clientCtx.Keyring.GetPubKey(name)
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

	sig, err := clientCtx.Keyring.Sign(name, signBz, f.SignMode())
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
