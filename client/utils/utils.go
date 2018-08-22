package utils

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	amino "github.com/tendermint/go-amino"
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

	txCtx, err = enrichCtxWithGasIfGasAuto(txCtx, cliCtx, cliCtx.FromAddressName, passphrase, msgs)
	if err != nil {
		return err
	}

	// build and sign the transaction
	txBytes, err := txCtx.BuildAndSign(cliCtx.FromAddressName, passphrase, msgs)
	if err != nil {
		return err
	}
	// broadcast to a Tendermint node
	return cliCtx.EnsureBroadcastTx(txBytes)
}

func enrichCtxWithGasIfGasAuto(txCtx authctx.TxContext, cliCtx context.CLIContext, name, passphrase string, msgs []sdk.Msg) (authctx.TxContext, error) {
	if cliCtx.Gas == 0 {
		return EnrichTxContextWithGas(txCtx, cliCtx, name, passphrase, msgs)
	}
	return txCtx, nil
}

// EnrichTxContextWithGas simulates the execution of a transaction to
// then populate the relevant TxContext.Gas field with the estimate
// obtained by the query.
func EnrichTxContextWithGas(txCtx authctx.TxContext, cliCtx context.CLIContext, name, passphrase string, msgs []sdk.Msg) (authctx.TxContext, error) {
	txCtxSimulation := txCtx.WithGas(0)
	txBytes, err := txCtxSimulation.BuildAndSign(name, passphrase, msgs)
	if err != nil {
		return txCtx, err
	}
	// run a simulation (via /app/simulate query) to
	// estimate gas and update TxContext accordingly
	rawRes, err := cliCtx.Query("/app/simulate", txBytes)
	if err != nil {
		return txCtx, err
	}
	estimate, err := parseQueryResponse(cliCtx.Codec, rawRes)
	if err != nil {
		return txCtx, err
	}
	adjusted := adjustGasEstimate(estimate, cliCtx.GasAdjustment)
	fmt.Fprintf(os.Stderr, "gas: [estimated = %v] [adjusted = %v]\n", estimate, adjusted)
	return txCtx.WithGas(adjusted), nil
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
