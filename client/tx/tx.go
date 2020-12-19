package tx

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// GenerateOrBroadcastTxCLI will either generate and print and unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxCLI(clientCtx client.Context, flagSet *pflag.FlagSet, msgs ...sdk.Msg) error {
	txf := NewFactoryCLI(clientCtx, flagSet)
	return GenerateOrBroadcastTxWithFactory(clientCtx, txf, msgs...)
}

// GenerateOrBroadcastTxWithFactory will either generate and print and unsigned transaction
// or sign it and broadcast it returning an error upon failure.
func GenerateOrBroadcastTxWithFactory(clientCtx client.Context, txf Factory, msgs ...sdk.Msg) error {
	if clientCtx.GenerateOnly {
		return GenerateTx(clientCtx, txf, msgs...)
	}

	return BroadcastTx(clientCtx, txf, msgs...)
}

// GenerateTx will generate an unsigned transaction and print it to the writer
// specified by ctx.Output. If simulation was requested, the gas will be
// simulated and also printed to the same writer before the transaction is
// printed.
func GenerateTx(clientCtx client.Context, txf Factory, msgs ...sdk.Msg) error {
	if txf.SimulateAndExecute() {
		if clientCtx.Offline {
			return errors.New("cannot estimate gas in offline mode")
		}

		_, adjusted, err := CalculateGas(clientCtx.QueryWithData, txf, msgs...)
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

	json, err := clientCtx.TxConfig.TxJSONEncoder()(tx.GetTx())
	if err != nil {
		return err
	}

	return clientCtx.PrintString(fmt.Sprintf("%s\n", json))
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
func BroadcastTx(clientCtx client.Context, txf Factory, msgs ...sdk.Msg) error {
	txf, err := PrepareFactory(clientCtx, txf)
	if err != nil {
		return err
	}

	if txf.SimulateAndExecute() || clientCtx.Simulate {
		_, adjusted, err := CalculateGas(clientCtx.QueryWithData, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", GasEstimateResponse{GasEstimate: txf.Gas()})
	}

	if clientCtx.Simulate {
		return nil
	}

	tx, err := BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return err
	}

	if !clientCtx.SkipConfirm {
		out, err := clientCtx.TxConfig.TxJSONEncoder()(tx.GetTx())
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

	err = Sign(txf, clientCtx.GetFromName(), tx, true)
	if err != nil {
		return err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(tx.GetTx())
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}

// WriteGeneratedTxResponse writes a generated unsigned transaction to the
// provided http.ResponseWriter. It will simulate gas costs if requested by the
// BaseReq. Upon any error, the error will be written to the http.ResponseWriter.
// Note that this function returns the legacy StdTx Amino JSON format for compatibility
// with legacy clients.
func WriteGeneratedTxResponse(
	ctx client.Context, w http.ResponseWriter, br rest.BaseReq, msgs ...sdk.Msg,
) {
	gasAdj, ok := rest.ParseFloat64OrReturnBadRequest(w, br.GasAdjustment, flags.DefaultGasAdjustment)
	if !ok {
		return
	}

	gasSetting, err := flags.ParseGasSetting(br.Gas)
	if rest.CheckBadRequestError(w, err) {
		return
	}

	txf := Factory{fees: br.Fees, gasPrices: br.GasPrices}.
		WithAccountNumber(br.AccountNumber).
		WithSequence(br.Sequence).
		WithGas(gasSetting.Gas).
		WithGasAdjustment(gasAdj).
		WithMemo(br.Memo).
		WithChainID(br.ChainID).
		WithSimulateAndExecute(br.Simulate).
		WithTxConfig(ctx.TxConfig).
		WithTimeoutHeight(br.TimeoutHeight)

	if br.Simulate || gasSetting.Simulate {
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
			rest.WriteSimulationResponse(w, ctx.LegacyAmino, txf.Gas())
			return
		}
	}

	tx, err := BuildUnsignedTx(txf, msgs...)
	if rest.CheckBadRequestError(w, err) {
		return
	}

	stdTx, err := ConvertTxToStdTx(ctx.LegacyAmino, tx.GetTx())
	if rest.CheckInternalServerError(w, err) {
		return
	}

	output, err := ctx.LegacyAmino.MarshalJSON(stdTx)
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
func BuildUnsignedTx(txf Factory, msgs ...sdk.Msg) (client.TxBuilder, error) {
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

	tx := txf.txConfig.NewTxBuilder()

	if err := tx.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	tx.SetMemo(txf.memo)
	tx.SetFeeAmount(fees)
	tx.SetGasLimit(txf.gas)
	tx.SetTimeoutHeight(txf.TimeoutHeight())

	return tx, nil
}

// BuildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be
// built.
func BuildSimTx(txf Factory, msgs ...sdk.Msg) ([]byte, error) {
	txb, err := BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return nil, err
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		PubKey: &secp256k1.PubKey{},
		Data: &signing.SingleSignatureData{
			SignMode: txf.signMode,
		},
		Sequence: txf.Sequence(),
	}
	if err := txb.SetSignatures(sig); err != nil {
		return nil, err
	}

	protoProvider, ok := txb.(authtx.ProtoTxProvider)
	if !ok {
		return nil, fmt.Errorf("cannot simulate amino tx")
	}
	simReq := tx.SimulateRequest{Tx: protoProvider.GetProtoTx()}

	return simReq.Marshal()
}

// CalculateGas simulates the execution of a transaction and returns the
// simulation response obtained by the query and the adjusted gas amount.
func CalculateGas(
	queryFunc func(string, []byte) ([]byte, int64, error), txf Factory, msgs ...sdk.Msg,
) (tx.SimulateResponse, uint64, error) {
	txBytes, err := BuildSimTx(txf, msgs...)
	if err != nil {
		return tx.SimulateResponse{}, 0, err
	}

	// TODO This should use the generated tx service Client.
	// https://github.com/cosmos/cosmos-sdk/issues/7726
	bz, _, err := queryFunc("/cosmos.tx.v1beta1.Service/Simulate", txBytes)
	if err != nil {
		return tx.SimulateResponse{}, 0, err
	}

	var simRes tx.SimulateResponse

	if err := simRes.Unmarshal(bz); err != nil {
		return tx.SimulateResponse{}, 0, err
	}

	return simRes, uint64(txf.GasAdjustment() * float64(simRes.GasInfo.GasUsed)), nil
}

// PrepareFactory ensures the account defined by ctx.GetFromAddress() exists and
// if the account number and/or the account sequence number are zero (not set),
// they will be queried for and set on the provided Factory. A new Factory with
// the updated fields will be returned.
func PrepareFactory(clientCtx client.Context, txf Factory) (Factory, error) {
	from := clientCtx.GetFromAddress()

	if err := txf.accountRetriever.EnsureExists(clientCtx, from); err != nil {
		return txf, err
	}

	initNum, initSeq := txf.accountNumber, txf.sequence
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.accountRetriever.GetAccountNumberSequence(clientCtx, from)
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

// SignWithPrivKey signs a given tx with the given private key, and returns the
// corresponding SignatureV2 if the signing is successful.
func SignWithPrivKey(
	signMode signing.SignMode, signerData authsigning.SignerData,
	txBuilder client.TxBuilder, priv cryptotypes.PrivKey, txConfig client.TxConfig,
	accSeq uint64,
) (signing.SignatureV2, error) {
	var sigV2 signing.SignatureV2

	// Generate the bytes to be signed.
	signBytes, err := txConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return sigV2, err
	}

	// Sign those bytes
	signature, err := priv.Sign(signBytes)
	if err != nil {
		return sigV2, err
	}

	// Construct the SignatureV2 struct
	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: signature,
	}

	sigV2 = signing.SignatureV2{
		PubKey:   priv.PubKey(),
		Data:     &sigData,
		Sequence: accSeq,
	}

	return sigV2, nil
}

func checkMultipleSigners(mode signing.SignMode, tx authsigning.Tx) error {
	if mode == signing.SignMode_SIGN_MODE_DIRECT &&
		len(tx.GetSigners()) > 1 {
		return sdkerrors.Wrap(sdkerrors.ErrNotSupported, "Signing in DIRECT mode is only supported for transactions with one signer only")
	}
	return nil
}

// Sign signs a given tx with a named key. The bytes signed over are canconical.
// The resulting signature will be added to the transaction builder overwriting the previous
// ones if overwrite=true (otherwise, the signature will be appended).
// Signing a transaction with mutltiple signers in the DIRECT mode is not supprted and will
// return an error.
// An error is returned upon failure.
func Sign(txf Factory, name string, txBuilder client.TxBuilder, overwriteSig bool) error {
	if txf.keybase == nil {
		return errors.New("keybase must be set prior to signing a transaction")
	}

	signMode := txf.signMode
	if signMode == signing.SignMode_SIGN_MODE_UNSPECIFIED {
		// use the SignModeHandler's default mode if unspecified
		signMode = txf.txConfig.SignModeHandler().DefaultMode()
	}
	if err := checkMultipleSigners(signMode, txBuilder.GetTx()); err != nil {
		return err
	}

	key, err := txf.keybase.Key(name)
	if err != nil {
		return err
	}
	pubKey := key.GetPubKey()
	signerData := authsigning.SignerData{
		ChainID:       txf.chainID,
		AccountNumber: txf.accountNumber,
		Sequence:      txf.sequence,
	}

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generated the
	// sign bytes. This is the reason for setting SetSignatures here, with a
	// nil signature.
	//
	// Note: this line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}
	var prevSignatures []signing.SignatureV2
	if !overwriteSig {
		prevSignatures, err = txBuilder.GetTx().GetSignaturesV2()
		if err != nil {
			return err
		}
	}
	if err := txBuilder.SetSignatures(sig); err != nil {
		return err
	}

	// Generate the bytes to be signed.
	bytesToSign, err := txf.txConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return err
	}

	// Sign those bytes
	sigBytes, _, err := txf.keybase.Sign(name, bytesToSign)
	if err != nil {
		return err
	}

	// Construct the SignatureV2 struct
	sigData = signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: sigBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}

	if overwriteSig {
		return txBuilder.SetSignatures(sig)
	}
	prevSignatures = append(prevSignatures, sig)
	return txBuilder.SetSignatures(prevSignatures...)
}

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}
