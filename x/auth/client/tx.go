package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Codec defines the x/auth account codec to be used for use with the
// AccountRetriever. The application must be sure to set this to their respective
// codec that implements the Codec interface and must be the same codec that
// passed to the x/auth module.
//
// TODO:/XXX: Using a package-level global isn't ideal and we should consider
// refactoring the module manager to allow passing in the correct module codec.
var Codec codec.Marshaler

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}

// GenerateOrBroadcastMsgs creates a StdTx given a series of messages. If
// the provided context has generate-only enabled, the tx will only be printed
// to STDOUT in a fully offline manner. Otherwise, the tx will be signed and
// broadcasted.
func GenerateOrBroadcastMsgs(clientCtx client.Context, txBldr authtypes.TxBuilder, msgs []sdk.Msg) error {
	if clientCtx.GenerateOnly {
		return PrintUnsignedStdTx(txBldr, clientCtx, msgs)
	}

	return CompleteAndBroadcastTxCLI(txBldr, clientCtx, msgs)
}

// CompleteAndBroadcastTxCLI implements a utility function that facilitates
// sending a series of messages in a signed transaction given a TxBuilder and a
// QueryContext. It ensures that the account exists, has a proper number and
// sequence set. In addition, it builds and signs a transaction with the
// supplied messages. Finally, it broadcasts the signed transaction to a node.
func CompleteAndBroadcastTxCLI(txBldr authtypes.TxBuilder, clientCtx client.Context, msgs []sdk.Msg) error {
	txBldr, err := PrepareTxBuilder(txBldr, clientCtx)
	if err != nil {
		return err
	}

	fromName := clientCtx.GetFromName()

	if txBldr.SimulateAndExecute() || clientCtx.Simulate {
		txBldr, err = EnrichWithGas(txBldr, clientCtx, msgs)
		if err != nil {
			return err
		}

		gasEst := GasEstimateResponse{GasEstimate: txBldr.Gas()}
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", gasEst.String())
	}

	if clientCtx.Simulate {
		return nil
	}

	if !clientCtx.SkipConfirm {
		stdSignMsg, err := txBldr.BuildSignMsg(msgs)
		if err != nil {
			return err
		}

		json := clientCtx.JSONMarshaler.MustMarshalJSON(stdSignMsg)
		if viper.GetBool(flags.FlagIndentResponse) {
			json, err = codec.MarshalIndentFromJSON(json)
			if err != nil {
				panic(err)
			}
		}

		_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", json)

		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stderr)
		if err != nil || !ok {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", "cancelled transaction")
			return err
		}
	}

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return clientCtx.PrintOutput(res)
}

// EnrichWithGas calculates the gas estimate that would be consumed by the
// transaction and set the transaction's respective value accordingly.
func EnrichWithGas(txBldr authtypes.TxBuilder, clientCtx client.Context, msgs []sdk.Msg) (authtypes.TxBuilder, error) {
	_, adjusted, err := simulateMsgs(txBldr, clientCtx, msgs)
	if err != nil {
		return txBldr, err
	}

	return txBldr.WithGas(adjusted), nil
}

// CalculateGas simulates the execution of a transaction and returns
// the simulation response obtained by the query and the adjusted gas amount.
func CalculateGas(
	queryFunc func(string, []byte) ([]byte, int64, error), cdc *codec.Codec,
	txBytes []byte, adjustment float64,
) (sdk.SimulationResponse, uint64, error) {

	// run a simulation (via /app/simulate query) to
	// estimate gas and update TxBuilder accordingly
	rawRes, _, err := queryFunc("/app/simulate", txBytes)
	if err != nil {
		return sdk.SimulationResponse{}, 0, err
	}

	simRes, err := parseQueryResponse(rawRes)
	if err != nil {
		return sdk.SimulationResponse{}, 0, err
	}

	adjusted := adjustGasEstimate(simRes.GasUsed, adjustment)
	return simRes, adjusted, nil
}

// PrintUnsignedStdTx builds an unsigned StdTx and prints it to os.Stdout.
func PrintUnsignedStdTx(txBldr authtypes.TxBuilder, clientCtx client.Context, msgs []sdk.Msg) error {
	stdTx, err := buildUnsignedStdTxOffline(txBldr, clientCtx, msgs)
	if err != nil {
		return err
	}

	json, err := clientCtx.JSONMarshaler.MarshalJSON(stdTx)
	if err != nil {
		return err
	}

	if viper.GetBool(flags.FlagIndentResponse) {
		json, err = codec.MarshalIndentFromJSON(json)
		if err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(clientCtx.Output, "%s\n", json)
	return nil
}

// SignStdTx appends a signature to a StdTx and returns a copy of it. If appendSig
// is false, it replaces the signatures already attached with the new signature.
// Don't perform online validation or lookups if offline is true.
func SignStdTx(
	txBldr authtypes.TxBuilder, clientCtx client.Context, name string,
	stdTx authtypes.StdTx, appendSig bool, offline bool,
) (authtypes.StdTx, error) {

	var signedStdTx authtypes.StdTx

	info, err := txBldr.Keybase().Key(name)
	if err != nil {
		return signedStdTx, err
	}

	addr := info.GetPubKey().Address()

	// check whether the address is a signer
	if !isTxSigner(sdk.AccAddress(addr), stdTx.GetSigners()) {
		return signedStdTx, fmt.Errorf("%s: %s", sdkerrors.ErrorInvalidSigner, name)
	}

	if !offline {
		txBldr, err = populateAccountFromState(txBldr, clientCtx, sdk.AccAddress(addr))
		if err != nil {
			return signedStdTx, err
		}
	}

	return txBldr.SignStdTx(name, stdTx, appendSig)
}

// SignStdTxWithSignerAddress attaches a signature to a StdTx and returns a copy of a it.
// Don't perform online validation or lookups if offline is true, else
// populate account and sequence numbers from a foreign account.
func SignStdTxWithSignerAddress(
	txBldr authtypes.TxBuilder, clientCtx client.Context,
	addr sdk.AccAddress, name string, stdTx authtypes.StdTx, offline bool,
) (signedStdTx authtypes.StdTx, err error) {

	// check whether the address is a signer
	if !isTxSigner(addr, stdTx.GetSigners()) {
		return signedStdTx, fmt.Errorf("%s: %s", sdkerrors.ErrorInvalidSigner, name)
	}

	if !offline {
		txBldr, err = populateAccountFromState(txBldr, clientCtx, addr)
		if err != nil {
			return signedStdTx, err
		}
	}

	return txBldr.SignStdTx(name, stdTx, false)
}

// Read and decode a StdTx from the given filename.  Can pass "-" to read from stdin.
func ReadTxFromFile(ctx client.Context, filename string) (tx sdk.Tx, err error) {
	var bytes []byte

	if filename == "-" {
		bytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		bytes, err = ioutil.ReadFile(filename)
	}

	if err != nil {
		return
	}

	return ctx.TxGenerator.TxJSONDecoder()(bytes)
}

// NewBatchScanner returns a new BatchScanner to read newline-delimited StdTx transactions from r.
func NewBatchScanner(cdc *codec.Codec, r io.Reader) *BatchScanner {
	return &BatchScanner{Scanner: bufio.NewScanner(r), cdc: cdc}
}

// BatchScanner provides a convenient interface for reading batch data such as a file
// of newline-delimited JSON encoded StdTx.
type BatchScanner struct {
	*bufio.Scanner
	stdTx        authtypes.StdTx
	cdc          *codec.Codec
	unmarshalErr error
}

// StdTx returns the most recent StdTx unmarshalled by a call to Scan.
func (bs BatchScanner) StdTx() authtypes.StdTx { return bs.stdTx }

// UnmarshalErr returns the first unmarshalling error that was encountered by the scanner.
func (bs BatchScanner) UnmarshalErr() error { return bs.unmarshalErr }

// Scan advances the Scanner to the next line.
func (bs *BatchScanner) Scan() bool {
	if !bs.Scanner.Scan() {
		return false
	}

	if err := bs.cdc.UnmarshalJSON(bs.Bytes(), &bs.stdTx); err != nil && bs.unmarshalErr == nil {
		bs.unmarshalErr = err
		return false
	}

	return true
}

func populateAccountFromState(
	txBldr authtypes.TxBuilder, clientCtx client.Context, addr sdk.AccAddress,
) (authtypes.TxBuilder, error) {

	num, seq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, addr)
	if err != nil {
		return txBldr, err
	}

	return txBldr.WithAccountNumber(num).WithSequence(seq), nil
}

// GetTxEncoder return tx encoder from global sdk configuration if ones is defined.
// Otherwise returns encoder with default logic.
func GetTxEncoder(cdc *codec.Codec) (encoder sdk.TxEncoder) {
	encoder = sdk.GetConfig().GetTxEncoder()
	if encoder == nil {
		encoder = authtypes.DefaultTxEncoder(cdc)
	}

	return encoder
}

// simulateMsgs simulates the transaction and returns the simulation response and
// the adjusted gas value.
func simulateMsgs(txBldr authtypes.TxBuilder, clientCtx client.Context, msgs []sdk.Msg) (sdk.SimulationResponse, uint64, error) {
	txBytes, err := txBldr.BuildTxForSim(msgs)
	if err != nil {
		return sdk.SimulationResponse{}, 0, err
	}

	return CalculateGas(clientCtx.QueryWithData, clientCtx.Codec, txBytes, txBldr.GasAdjustment())
}

func adjustGasEstimate(estimate uint64, adjustment float64) uint64 {
	return uint64(adjustment * float64(estimate))
}

func parseQueryResponse(bz []byte) (sdk.SimulationResponse, error) {
	var simRes sdk.SimulationResponse
	if err := jsonpb.Unmarshal(strings.NewReader(string(bz)), &simRes); err != nil {
		return sdk.SimulationResponse{}, err
	}

	return simRes, nil
}

// PrepareTxBuilder populates a TxBuilder in preparation for the build of a Tx.
func PrepareTxBuilder(txBldr authtypes.TxBuilder, clientCtx client.Context) (authtypes.TxBuilder, error) {
	from := clientCtx.GetFromAddress()
	accGetter := clientCtx.AccountRetriever
	if err := accGetter.EnsureExists(clientCtx, from); err != nil {
		return txBldr, err
	}

	txbldrAccNum, txbldrAccSeq := txBldr.AccountNumber(), txBldr.Sequence()
	// TODO: (ref #1903) Allow for user supplied account number without
	// automatically doing a manual lookup.
	if txbldrAccNum == 0 || txbldrAccSeq == 0 {
		num, seq, err := accGetter.GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return txBldr, err
		}

		if txbldrAccNum == 0 {
			txBldr = txBldr.WithAccountNumber(num)
		}
		if txbldrAccSeq == 0 {
			txBldr = txBldr.WithSequence(seq)
		}
	}

	return txBldr, nil
}

func buildUnsignedStdTxOffline(txBldr authtypes.TxBuilder, clientCtx client.Context, msgs []sdk.Msg) (stdTx authtypes.StdTx, err error) {
	if txBldr.SimulateAndExecute() {
		if clientCtx.Offline {
			return stdTx, errors.New("cannot estimate gas in offline mode")
		}

		txBldr, err = EnrichWithGas(txBldr, clientCtx, msgs)
		if err != nil {
			return stdTx, err
		}

		_, _ = fmt.Fprintf(os.Stderr, "estimated gas = %v\n", txBldr.Gas())
	}

	stdSignMsg, err := txBldr.BuildSignMsg(msgs)
	if err != nil {
		return stdTx, err
	}

	return authtypes.NewStdTx(stdSignMsg.Msgs, stdSignMsg.Fee, nil, stdSignMsg.Memo), nil
}

func isTxSigner(user sdk.AccAddress, signers []sdk.AccAddress) bool {
	for _, s := range signers {
		if bytes.Equal(user.Bytes(), s.Bytes()) {
			return true
		}
	}

	return false
}
