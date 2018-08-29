package utils

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/common"
)

// DefaultGasAdjustment is applied to gas estimates to avoid tx
// execution failures due to state changes that might
// occur between the tx simulation and the actual run.
const DefaultGasAdjustment = 1.2

// SendTx implements a auxiliary handler that facilitates sending a series of
// messages in a signed transaction given a TxContext and a QueryContext. It
// ensures that the account exists, has a proper number and sequence set. In
// addition, it builds and signs a transaction with the supplied messages.
// Finally, it broadcasts the signed transaction to a node.
func SendTx(txCtx authctx.TxContext, cliCtx context.CLIContext, msgs []sdk.Msg) error {
	if err := cliCtx.EnsureAccountExists(); err != nil {
		return err
	}

	from, err := cliCtx.GetFromAddress()
	if err != nil {
		return err
	}

	// TODO: (ref #1903) Allow for user supplied account number without
	// automatically doing a manual lookup.
	if txCtx.AccountNumber == 0 {
		accNum, err := cliCtx.GetAccountNumber(from)
		if err != nil {
			return err
		}

		txCtx = txCtx.WithAccountNumber(accNum)
	}

	// TODO: (ref #1903) Allow for user supplied account sequence without
	// automatically doing a manual lookup.
	if txCtx.Sequence == 0 {
		accSeq, err := cliCtx.GetAccountSequence(from)
		if err != nil {
			return err
		}

		txCtx = txCtx.WithSequence(accSeq)
	}

	passphrase, err := keys.GetPassphrase(cliCtx.FromAddressName)
	if err != nil {
		return err
	}

	if cliCtx.Gas == 0 {
		txCtx, err = EnrichCtxWithGas(txCtx, cliCtx, cliCtx.FromAddressName, passphrase, msgs)
		if err != nil {
			return err
		}
	}

	// build and sign the transaction
	txBytes, err := txCtx.BuildAndSign(cliCtx.FromAddressName, passphrase, msgs)
	if err != nil {
		return err
	}
	// broadcast to a Tendermint node
	return cliCtx.EnsureBroadcastTx(txBytes)
}

// EnrichCtxWithGas calculates the gas estimate that would be consumed by the
// transaction and set the transaction's respective value accordingly.
func EnrichCtxWithGas(txCtx authctx.TxContext, cliCtx context.CLIContext, name, passphrase string, msgs []sdk.Msg) (authctx.TxContext, error) {
	txBytes, err := BuildAndSignTxWithZeroGas(txCtx, name, passphrase, msgs)
	if err != nil {
		return txCtx, err
	}
	estimate, adjusted, err := CalculateGas(cliCtx.Query, cliCtx.Codec, txBytes, cliCtx.GasAdjustment)
	if err != nil {
		return txCtx, err
	}
	fmt.Fprintf(os.Stderr, "gas: [estimated = %v] [adjusted = %v]\n", estimate, adjusted)
	return txCtx.WithGas(adjusted), nil
}

// BuildAndSignTxWithZeroGas builds transactions with GasWanted set to 0.
func BuildAndSignTxWithZeroGas(txCtx authctx.TxContext, name, passphrase string, msgs []sdk.Msg) ([]byte, error) {
	return txCtx.WithGas(0).BuildAndSign(name, passphrase, msgs)
}

// CalculateGas simulates the execution of a transaction and returns
// both the estimate obtained by the query and the adjusted amount.
func CalculateGas(queryFunc func(string, common.HexBytes) ([]byte, error), cdc *amino.Codec, txBytes []byte, adjustment float64) (estimate, adjusted int64, err error) {
	// run a simulation (via /app/simulate query) to
	// estimate gas and update TxContext accordingly
	rawRes, err := queryFunc("/app/simulate", txBytes)
	if err != nil {
		return
	}
	estimate, err = parseQueryResponse(cdc, rawRes)
	if err != nil {
		return
	}
	adjusted = adjustGasEstimate(estimate, adjustment)
	fmt.Fprintf(os.Stderr, "gas: [estimated = %v] [adjusted = %v]\n", estimate, adjusted)
	return
}

func adjustGasEstimate(estimate int64, adjustment float64) int64 {
	if adjustment == 0 {
		return int64(DefaultGasAdjustment * float64(estimate))
	}
	return int64(adjustment * float64(estimate))
}

func parseQueryResponse(cdc *amino.Codec, rawRes []byte) (int64, error) {
	var simulationResult sdk.Result
	if err := cdc.UnmarshalBinary(rawRes, &simulationResult); err != nil {
		return 0, err
	}
	return simulationResult.GasUsed, nil
}
