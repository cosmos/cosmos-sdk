package context

import (
	"bytes"
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// CompleteAndBroadcastTxCli implements a utility function that facilitates
// sending a series of messages in a signed transaction given a TxBuilder and a
// QueryContext. It ensures that the account exists, has a proper number and
// sequence set. In addition, it builds and signs a transaction with the
// supplied messages. Finally, it broadcasts the signed transaction to a node.
//
// NOTE: Also see CompleteAndBroadcastTxREST.
func (ctx CLIContext) CompleteAndBroadcastTxCli(msgs []sdk.Msg) error {
	ctx, err := ctx.WithTxBldrAddress(ctx.FromAddr())
	if err != nil {
		return err
	}

	if ctx.TxBldr.GetSimulateAndExecute() || ctx.Simulate {
		ctx, err = ctx.EnrichCtxWithGas(ctx.FromName(), msgs)
		if err != nil {
			return err
		}
	}

	if ctx.Simulate {
		return nil
	}

	passphrase, err := keys.GetPassphrase(ctx.FromName())
	if err != nil {
		return err
	}

	// build and sign the transaction
	txBytes, err := ctx.TxBldr.BuildAndSign(ctx.FromName(), passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	_, err = ctx.BroadcastTx(txBytes)
	return err
}

// EnrichCtxWithGas calculates the gas estimate that would be consumed by the
// transaction and set the transaction's respective value accordingly.
func (ctx CLIContext) EnrichCtxWithGas(name string, msgs []sdk.Msg) (CLIContext, error) {
	_, adjusted, err := ctx.simulateMsgs(name, msgs)
	if err != nil {
		return ctx, err
	}
	ctx.TxBldr = ctx.TxBldr.WithGas(adjusted)
	return ctx, nil
}

// CalculateGas simulates the execution of a transaction and returns
// both the estimate obtained by the query and the adjusted amount.
func CalculateGas(queryFunc func(string, common.HexBytes) ([]byte, error), cdc *amino.Codec, txBytes []byte, adjustment float64) (uint64, uint64, error) {

	rawRes, err := queryFunc("/app/simulate", txBytes)
	if err != nil {
		return 0, 0, err
	}

	var simulationResult sdk.Result

	if err = cdc.UnmarshalBinaryLengthPrefixed(rawRes, &simulationResult); err != nil {
		return 0, 0, err
	}

	est := simulationResult.GasUsed

	return est, uint64(adjustment * float64(est)), nil
}

// PrintUnsignedStdTx builds an unsigned StdTx and prints it to ctx.Ouput.
// Doesn't perform online validation or lookups if offline is true.
func (ctx CLIContext) PrintUnsignedStdTx(w io.Writer, msgs []sdk.Msg, offline bool) (err error) {

	var stdTx auth.StdTx

	if offline {
		stdTx, err = ctx.buildUnsignedStdTxOffline(msgs)
	} else {
		stdTx, err = ctx.buildUnsignedStdTx(msgs)
	}

	if err != nil {
		return
	}

	if json, err := ctx.Codec.MarshalJSON(stdTx); err == nil {
		fmt.Fprintf(w, "%s\n", json)
	}

	return
}

// SignStdTx appends a signature to a StdTx and returns a copy of a it. If appendSig
// is false, it replaces the signatures already attached with the new signature.
// Doesn't perform online validation or lookups if offline is true.
func (ctx CLIContext) SignStdTx(name string, stdTx auth.StdTx, appendSig bool, offline bool) (auth.StdTx, error) {
	var signedStdTx auth.StdTx

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return signedStdTx, err
	}

	info, err := keybase.Get(name)
	if err != nil {
		return signedStdTx, err
	}

	addr := sdk.AccAddress(info.GetPubKey().Address())

	// check whether the address is a signer
	if !isTxSigner(addr, stdTx.GetSigners()) {
		return signedStdTx, ErrInvalidSigner(addr, stdTx.GetSigners())
	}

	if !offline {
		ctx, err = ctx.WithTxBldrAddress(addr)
		if err != nil {
			return signedStdTx, err
		}
	}

	passphrase, err := keys.GetPassphrase(name)
	if err != nil {
		return signedStdTx, err
	}

	return ctx.TxBldr.SignStdTx(name, passphrase, stdTx, appendSig)
}

// SignStdTxWithSignerAddress attaches a signature to a StdTx and returns a copy of a it.
// Doesn't perform online validation or lookups if offline is true, else
// populate account and sequence numbers from a foreign account.
func (ctx CLIContext) SignStdTxWithSignerAddress(addr sdk.AccAddress,
	name string, stdTx auth.StdTx, offline bool) (signedStdTx auth.StdTx, err error) {

	// check whether the address is a signer
	if !isTxSigner(addr, stdTx.GetSigners()) {
		return signedStdTx, ErrInvalidSigner(addr, stdTx.GetSigners())
	}

	if !offline {
		ctx, err = ctx.WithTxBldrAddress(addr)
		if err != nil {
			return signedStdTx, err
		}
	}

	passphrase, err := keys.GetPassphrase(name)
	if err != nil {
		return signedStdTx, err
	}

	return ctx.TxBldr.SignStdTx(name, passphrase, stdTx, false)
}

// GetTxEncoder return tx encoder from global sdk configuration if ones is defined.
// Otherwise returns encoder with default logic.
func GetTxEncoder(cdc *codec.Codec) (encoder sdk.TxEncoder) {
	encoder = sdk.GetConfig().GetTxEncoder()
	if encoder == nil {
		encoder = auth.DefaultTxEncoder(cdc)
	}
	return
}

// nolint
// SimulateMsgs simulates the transaction and returns the gas estimate and the adjusted value.
func (ctx CLIContext) simulateMsgs(name string, msgs []sdk.Msg) (uint64, uint64, error) {
	txBytes, err := ctx.TxBldr.BuildWithPubKey(name, msgs)
	if err != nil {
		return 0, 0, err
	}
	adjustment := ctx.TxBldr.GetGasAdjustment()
	return CalculateGas(ctx.Query, ctx.Codec, txBytes, adjustment)
}

// buildUnsignedStdTx builds a StdTx as per the parameters passed in the
// contexts. Gas is automatically estimated if gas wanted is set to 0.
func (ctx CLIContext) buildUnsignedStdTx(msgs []sdk.Msg) (stdTx auth.StdTx, err error) {
	ctx, err = ctx.WithTxBldrAddress(ctx.FromAddr())
	if err != nil {
		return
	}
	return ctx.buildUnsignedStdTxOffline(msgs)
}

func (ctx CLIContext) buildUnsignedStdTxOffline(msgs []sdk.Msg) (stdTx auth.StdTx, err error) {
	if ctx.TxBldr.GetSimulateAndExecute() {
		var name = ctx.FromName()

		ctx, err = ctx.EnrichCtxWithGas(name, msgs)
		if err != nil {
			return
		}
		// fmt.Fprintf(os.Stderr, "estimated gas = %v\n", txBldr.GetGas())
	}

	stdSignMsg, err := ctx.TxBldr.Build(msgs)
	if err != nil {
		return
	}
	return auth.NewStdTx(stdSignMsg.Msgs, stdSignMsg.Fee, nil, stdSignMsg.Memo), nil
}

func isTxSigner(user sdk.AccAddress, signers []sdk.AccAddress) bool {
	for _, s := range signers {
		if bytes.Equal(user.Bytes(), s.Bytes()) {
			return true
		}
	}
	return false
}
