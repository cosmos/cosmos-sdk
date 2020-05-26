package tx

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gogo/protobuf/jsonpb"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// GenerateOrBroadcastTx will either generate and print and unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTx(ctx context.CLIContext, msgs ...sdk.Msg) error {
	txf := NewFactoryFromCLI(ctx.Input).WithTxGenerator(ctx.TxGenerator).WithAccountRetriever(ctx.AccountRetriever)
	return GenerateOrBroadcastTxWithFactory(ctx, txf, msgs...)
}

// GenerateOrBroadcastTxWithFactory will either generate and print and unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxWithFactory(ctx context.CLIContext, txf Factory, msgs ...sdk.Msg) error {
	if ctx.GenerateOnly {
		return GenerateTx(ctx, txf, msgs...)
	}

	return BroadcastTx(ctx, txf, msgs...)
}

// GenerateTx will generate an unsigned transaction and print it to the writer
// specified by ctx.Output. If simulation was requested, the gas will be
// simulated and also printed to the same writer before the transaction is
// printed.
func GenerateTx(ctx context.CLIContext, txf Factory, msgs ...sdk.Msg) error {
	if txf.SimulateAndExecute() {
		if ctx.Offline {
			return errors.New("cannot estimate gas in offline mode")
		}

		_, adjusted, err := CalculateGas(ctx.QueryWithData, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: txf.Gas()})
	}

	tx, err := BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return err
	}

	return ctx.Println(tx.GetTx())
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func BroadcastTx(ctx context.CLIContext, txf Factory, msgs ...sdk.Msg) error {
	txf, err := PrepareFactory(ctx, txf)
	if err != nil {
		return err
	}

	if txf.SimulateAndExecute() || ctx.Simulate {
		_, adjusted, err := CalculateGas(ctx.QueryWithData, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: txf.Gas()})
	}

	if ctx.Simulate {
		return nil
	}

	tx, err := BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return err
	}

	if !ctx.SkipConfirm {
		out, err := ctx.JSONMarshaler.MarshalJSON(tx)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", out)

		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stderr)

		if err != nil || !ok {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", "cancelled transaction")
			return err
		}
	}

	txBytes, err := Sign(txf, ctx.GetFromName(), clientkeys.DefaultKeyPass, tx)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return ctx.Println(res)
}

// WriteGeneratedTxResponse writes a generated unsigned transaction to the
// provided http.ResponseWriter. It will simulate gas costs if requested by the
// BaseReq. Upon any error, the error will be written to the http.ResponseWriter.
func WriteGeneratedTxResponse(
	ctx context.CLIContext, w http.ResponseWriter, br rest.BaseReq, msgs ...sdk.Msg,
) {
	gasAdj, ok := rest.ParseFloat64OrReturnBadRequest(w, br.GasAdjustment, flags.DefaultGasAdjustment)
	if !ok {
		return
	}

	simAndExec, gas, err := flags.ParseGas(br.Gas)
	if rest.CheckBadRequestError(w, err) {
		return
	}

	txf := Factory{fees: br.Fees, gasPrices: br.GasPrices}.
		WithAccountNumber(br.AccountNumber).
		WithSequence(br.Sequence).
		WithGas(gas).
		WithGasAdjustment(gasAdj).
		WithMemo(br.Memo).
		WithChainID(br.ChainID).
		WithSimulateAndExecute(br.Simulate).
		WithTxGenerator(ctx.TxGenerator)

	if br.Simulate || simAndExec {
		if gasAdj < 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.ErrorInvalidGasAdjustment.Error())
			return
		}

		_, adjusted, err := CalculateGas(ctx.QueryWithData, txf, msgs...)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		txf = txf.WithGas(adjusted)

		if br.Simulate {
			rest.WriteSimulationResponse(w, ctx.JSONMarshaler, txf.Gas())
			return
		}
	}

	tx, err := BuildUnsignedTx(txf, msgs...)
	if rest.CheckBadRequestError(w, err) {
		return
	}

	output, err := ctx.JSONMarshaler.MarshalJSON(tx)
	if rest.CheckInternalServerError(w, err) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
}

// BuildUnsignedTx builds a transaction to be signed given a set of messages. The
// transaction is initially created via the provided factory's generator. Once
// created, the fee, memo, and messages are set.
func BuildUnsignedTx(txf Factory, msgs ...sdk.Msg) (context.TxBuilder, error) {
	if txf.chainID == "" {
		return nil, fmt.Errorf("chain ID required but not specified")
	}

	fees := txf.fees

	if !txf.gasPrices.IsZero() {
		if !fees.IsZero() {
			return nil, errors.New("cannot provide both fees and gas prices")
		}

		glDec := sdk.NewDec(int64(txf.gas))

		// Derive the fees based on the provided gas prices, where
		// fee = ceil(gasPrice * gasLimit).
		fees = make(sdk.Coins, len(txf.gasPrices))

		for i, gp := range txf.gasPrices {
			fee := gp.Amount.Mul(glDec)
			fees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	clientFee := txf.txGenerator.NewFee()
	clientFee.SetAmount(fees)
	clientFee.SetGas(txf.gas)

	tx := txf.txGenerator.NewTx()
	tx.SetMemo(txf.memo)

	if err := tx.SetFee(clientFee); err != nil {
		return nil, err
	}

	if err := tx.SetSignatures(); err != nil {
		return nil, err
	}

	if err := tx.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	return tx, nil
}

// BuildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be
// built.
func BuildSimTx(txf Factory, msgs ...sdk.Msg) ([]byte, error) {
	tx, err := BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return nil, err
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := txf.txGenerator.NewSignature()

	if err := tx.SetSignatures(sig); err != nil {
		return nil, err
	}

	return txf.txGenerator.MarshalTx(tx.GetTx())
}

// CalculateGas simulates the execution of a transaction and returns the
// simulation response obtained by the query and the adjusted gas amount.
func CalculateGas(
	queryFunc func(string, []byte) ([]byte, int64, error), txf Factory, msgs ...sdk.Msg,
) (sdk.SimulationResponse, uint64, error) {
	txBytes, err := BuildSimTx(txf, msgs...)
	if err != nil {
		return sdk.SimulationResponse{}, 0, err
	}

	bz, _, err := queryFunc("/app/simulate", txBytes)
	if err != nil {
		return sdk.SimulationResponse{}, 0, err
	}

	var simRes sdk.SimulationResponse
	if err := jsonpb.Unmarshal(strings.NewReader(string(bz)), &simRes); err != nil {
		return sdk.SimulationResponse{}, 0, err
	}

	return simRes, uint64(txf.GasAdjustment() * float64(simRes.GasUsed)), nil
}

// PrepareFactory ensures the account defined by ctx.GetFromAddress() exists and
// if the account number and/or the account sequence number are zero (not set),
// they will be queried for and set on the provided Factory. A new Factory with
// the updated fields will be returned.
func PrepareFactory(ctx context.CLIContext, txf Factory) (Factory, error) {
	from := ctx.GetFromAddress()

	if err := txf.accountRetriever.EnsureExists(ctx, from); err != nil {
		return txf, err
	}

	initNum, initSeq := txf.accountNumber, txf.sequence
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.accountRetriever.GetAccountNumberSequence(ctx, from)
		if err != nil {
			return txf, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
		}
	}

	return txf, nil
}

// Sign signs a given tx with the provided name and passphrase. If the Factory's
// Keybase is not set, a new one will be created based on the client's backend.
// The bytes signed over are canconical. The resulting signature will be set on
// the transaction. Finally, the marshaled transaction is returned. An error is
// returned upon failure.
//
// Note, It is assumed the Factory has the necessary fields set that are required
// by the CanonicalSignBytes call.
func Sign(txf Factory, name, passphrase string, tx context.TxBuilder) ([]byte, error) {
	if txf.keybase == nil {
		return nil, errors.New("keybase must be set prior to signing a transaction")
	}

	signBytes, err := tx.CanonicalSignBytes(txf.chainID, txf.accountNumber, txf.sequence)
	if err != nil {
		return nil, err
	}

	sigBytes, pubkey, err := txf.keybase.Sign(name, signBytes)
	if err != nil {
		return nil, err
	}

	sig := txf.txGenerator.NewSignature()
	sig.SetSignature(sigBytes)

	if err := sig.SetPubKey(pubkey); err != nil {
		return nil, err
	}

	if err := tx.SetSignatures(sig); err != nil {
		return nil, err
	}

	return txf.txGenerator.MarshalTx(tx.GetTx())
}

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}
