package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

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
func GenerateOrBroadcastMsgs(cliCtx context.CLIContext, txBldr authtypes.TxBuilder, msgs []sdk.Msg) error {
	if cliCtx.GenerateOnly {
		return PrintUnsignedStdTx(txBldr, cliCtx, msgs)
	}

	return CompleteAndBroadcastTxCLI(txBldr, cliCtx, msgs)
}

// CompleteAndBroadcastTxCLI implements a utility function that facilitates
// sending a series of messages in a signed transaction given a TxBuilder and a
// QueryContext. It ensures that the account exists, has a proper number and
// sequence set. In addition, it builds and signs a transaction with the
// supplied messages. Finally, it broadcasts the signed transaction to a node.
func CompleteAndBroadcastTxCLI(txBldr authtypes.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg) error {
	txBldr, err := PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return err
	}

	fromName := cliCtx.GetFromName()

	if txBldr.SimulateAndExecute() || cliCtx.Simulate {
		txBldr, err = EnrichWithGas(txBldr, cliCtx, msgs)
		if err != nil {
			return err
		}

		gasEst := GasEstimateResponse{GasEstimate: txBldr.Gas()}
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", gasEst.String())
	}

	if cliCtx.Simulate {
		return nil
	}

	if !cliCtx.SkipConfirm {
		stdSignMsg, err := txBldr.BuildSignMsg(msgs)
		if err != nil {
			return err
		}

		var json []byte
		if viper.GetBool(flags.FlagIndentResponse) {
			json, err = cliCtx.Codec.MarshalJSONIndent(stdSignMsg, "", "  ")
			if err != nil {
				panic(err)
			}
		} else {
			json = cliCtx.Codec.MustMarshalJSON(stdSignMsg)
		}

		_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", json)

		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf)
		if err != nil || !ok {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", "cancelled transaction")
			return err
		}
	}

	passphrase, err := keys.GetPassphrase(fromName)
	if err != nil {
		return err
	}

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := cliCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return cliCtx.PrintOutput(res)
}

// EnrichWithGas calculates the gas estimate that would be consumed by the
// transaction and set the transaction's respective value accordingly.
func EnrichWithGas(txBldr authtypes.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg) (authtypes.TxBuilder, error) {
	_, adjusted, err := simulateMsgs(txBldr, cliCtx, msgs)
	if err != nil {
		return txBldr, err
	}

	return txBldr.WithGas(adjusted), nil
}

// CalculateGas simulates the execution of a transaction and returns
// both the estimate obtained by the query and the adjusted amount.
func CalculateGas(
	queryFunc func(string, []byte) ([]byte, int64, error), cdc *codec.Codec,
	txBytes []byte, adjustment float64,
) (estimate, adjusted uint64, err error) {

	// run a simulation (via /app/simulate query) to
	// estimate gas and update TxBuilder accordingly
	rawRes, _, err := queryFunc("/app/simulate", txBytes)
	if err != nil {
		return estimate, adjusted, err
	}

	estimate, err = parseQueryResponse(cdc, rawRes)
	if err != nil {
		return
	}

	adjusted = adjustGasEstimate(estimate, adjustment)
	return estimate, adjusted, nil
}

// PrintUnsignedStdTx builds an unsigned StdTx and prints it to os.Stdout.
func PrintUnsignedStdTx(txBldr authtypes.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg) error {
	stdTx, err := buildUnsignedStdTxOffline(txBldr, cliCtx, msgs)
	if err != nil {
		return err
	}

	json, err := cliCtx.Codec.MarshalJSON(stdTx)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cliCtx.Output, "%s\n", json)
	return nil
}

// SignStdTx appends a signature to a StdTx and returns a copy of it. If appendSig
// is false, it replaces the signatures already attached with the new signature.
// Don't perform online validation or lookups if offline is true.
func SignStdTx(
	txBldr authtypes.TxBuilder, cliCtx context.CLIContext, name string,
	stdTx authtypes.StdTx, appendSig bool, offline bool,
) (authtypes.StdTx, error) {

	var signedStdTx authtypes.StdTx

	info, err := txBldr.Keybase().Get(name)
	if err != nil {
		return signedStdTx, err
	}

	addr := info.GetPubKey().Address()

	// check whether the address is a signer
	if !isTxSigner(sdk.AccAddress(addr), stdTx.GetSigners()) {
		return signedStdTx, fmt.Errorf("%s: %s", errInvalidSigner, name)
	}

	if !offline {
		txBldr, err = populateAccountFromState(txBldr, cliCtx, sdk.AccAddress(addr))
		if err != nil {
			return signedStdTx, err
		}
	}

	passphrase, err := keys.GetPassphrase(name)
	if err != nil {
		return signedStdTx, err
	}

	return txBldr.SignStdTx(name, passphrase, stdTx, appendSig)
}

// SignStdTxWithSignerAddress attaches a signature to a StdTx and returns a copy of a it.
// Don't perform online validation or lookups if offline is true, else
// populate account and sequence numbers from a foreign account.
func SignStdTxWithSignerAddress(txBldr authtypes.TxBuilder, cliCtx context.CLIContext,
	addr sdk.AccAddress, name string, stdTx authtypes.StdTx,
	offline bool) (signedStdTx authtypes.StdTx, err error) {

	// check whether the address is a signer
	if !isTxSigner(addr, stdTx.GetSigners()) {
		return signedStdTx, fmt.Errorf("%s: %s", errInvalidSigner, name)
	}

	if !offline {
		txBldr, err = populateAccountFromState(txBldr, cliCtx, addr)
		if err != nil {
			return signedStdTx, err
		}
	}

	passphrase, err := keys.GetPassphrase(name)
	if err != nil {
		return signedStdTx, err
	}

	return txBldr.SignStdTx(name, passphrase, stdTx, false)
}

// Read and decode a StdTx from the given filename.  Can pass "-" to read from stdin.
func ReadStdTxFromFile(cdc *codec.Codec, filename string) (stdTx authtypes.StdTx, err error) {
	var bytes []byte

	if filename == "-" {
		bytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		bytes, err = ioutil.ReadFile(filename)
	}

	if err != nil {
		return
	}

	if err = cdc.UnmarshalJSON(bytes, &stdTx); err != nil {
		return
	}

	return
}

func populateAccountFromState(
	txBldr authtypes.TxBuilder, cliCtx context.CLIContext, addr sdk.AccAddress,
) (authtypes.TxBuilder, error) {

	num, seq, err := authtypes.NewAccountRetriever(cliCtx).GetAccountNumberSequence(addr)
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

// nolint
// SimulateMsgs simulates the transaction and returns the gas estimate and the adjusted value.
func simulateMsgs(txBldr authtypes.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg) (estimated, adjusted uint64, err error) {
	txBytes, err := txBldr.BuildTxForSim(msgs)
	if err != nil {
		return
	}

	estimated, adjusted, err = CalculateGas(cliCtx.QueryWithData, cliCtx.Codec, txBytes, txBldr.GasAdjustment())
	return
}

func adjustGasEstimate(estimate uint64, adjustment float64) uint64 {
	return uint64(adjustment * float64(estimate))
}

func parseQueryResponse(cdc *codec.Codec, rawRes []byte) (uint64, error) {
	var simulationResult sdk.Result
	if err := cdc.UnmarshalBinaryLengthPrefixed(rawRes, &simulationResult); err != nil {
		return 0, err
	}

	return simulationResult.GasUsed, nil
}

// PrepareTxBuilder populates a TxBuilder in preparation for the build of a Tx.
func PrepareTxBuilder(txBldr authtypes.TxBuilder, cliCtx context.CLIContext) (authtypes.TxBuilder, error) {
	from := cliCtx.GetFromAddress()

	accGetter := authtypes.NewAccountRetriever(cliCtx)
	if err := accGetter.EnsureExists(from); err != nil {
		return txBldr, err
	}

	txbldrAccNum, txbldrAccSeq := txBldr.AccountNumber(), txBldr.Sequence()
	// TODO: (ref #1903) Allow for user supplied account number without
	// automatically doing a manual lookup.
	if txbldrAccNum == 0 || txbldrAccSeq == 0 {
		num, seq, err := authtypes.NewAccountRetriever(cliCtx).GetAccountNumberSequence(from)
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

func buildUnsignedStdTxOffline(txBldr authtypes.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg) (stdTx authtypes.StdTx, err error) {
	if txBldr.SimulateAndExecute() {
		if cliCtx.GenerateOnly {
			return stdTx, errors.New("cannot estimate gas with generate-only")
		}

		txBldr, err = EnrichWithGas(txBldr, cliCtx, msgs)
		if err != nil {
			return stdTx, err
		}

		_, _ = fmt.Fprintf(os.Stderr, "estimated gas = %v\n", txBldr.Gas())
	}

	stdSignMsg, err := txBldr.BuildSignMsg(msgs)
	if err != nil {
		return stdTx, nil
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
