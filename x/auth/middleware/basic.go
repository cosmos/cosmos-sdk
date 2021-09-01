package middleware

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ValidateBasicDecorator will call tx.ValidateBasic, msg.ValidateBasic(for each msg inside tx)
// and return any non-nil error.
// If ValidateBasic passes, middleware calls next middleware in chain. Note,
// validateBasicMiddleware will not get executed on ReCheckTx since it
// is not dependent on application state.
type validateBasicMiddleware struct {
	next txtypes.Handler
}

func ValidateBasicMiddleware(txh txtypes.Handler) txtypes.Handler {
	return validateBasicMiddleware{
		next: txh,
	}
}

var _ txtypes.Handler = validateBasicMiddleware{}

// CheckTx implements tx.Handler.CheckTx.
func (basic validateBasicMiddleware) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// no need to validate basic on recheck tx, call next antehandler
	if sdkCtx.IsReCheckTx() {
		return basic.next.CheckTx(ctx, tx, req)
	}

	if err := tx.ValidateBasic(); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return basic.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (basic validateBasicMiddleware) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := tx.ValidateBasic(); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return basic.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (basic validateBasicMiddleware) SimulateTx(ctx context.Context, tx sdk.Tx, req txtypes.RequestSimulateTx) (txtypes.ResponseSimulateTx, error) {
	if err := tx.ValidateBasic(); err != nil {
		return txtypes.ResponseSimulateTx{}, err
	}

	return basic.next.SimulateTx(ctx, tx, req)
}

var _ txtypes.Handler = txTimeoutHeightMiddleware{}

type (
	// TxTimeoutHeightMiddleware defines an AnteHandler decorator that checks for a
	// tx height timeout.
	txTimeoutHeightMiddleware struct {
		next txtypes.Handler
	}

	// TxWithTimeoutHeight defines the interface a tx must implement in order for
	// TxHeightTimeoutDecorator to process the tx.
	TxWithTimeoutHeight interface {
		sdk.Tx

		GetTimeoutHeight() uint64
	}
)

// TxTimeoutHeightMiddleware defines an AnteHandler decorator that checks for a
// tx height timeout.
func TxTimeoutHeightMiddleware(txh txtypes.Handler) txtypes.Handler {
	return txTimeoutHeightMiddleware{
		next: txh,
	}
}

// CheckTx implements tx.Handler.CheckTx.
func (txh txTimeoutHeightMiddleware) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	timeoutTx, ok := tx.(TxWithTimeoutHeight)
	if !ok {
		return abci.ResponseCheckTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "expected tx to implement TxWithTimeoutHeight")
	}

	timeoutHeight := timeoutTx.GetTimeoutHeight()
	if timeoutHeight > 0 && uint64(sdkCtx.BlockHeight()) > timeoutHeight {
		return abci.ResponseCheckTx{}, sdkerrors.Wrapf(
			sdkerrors.ErrTxTimeoutHeight, "block height: %d, timeout height: %d", sdkCtx.BlockHeight(), timeoutHeight,
		)
	}

	return txh.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh txTimeoutHeightMiddleware) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	timeoutTx, ok := tx.(TxWithTimeoutHeight)
	if !ok {
		return abci.ResponseDeliverTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "expected tx to implement TxWithTimeoutHeight")
	}

	timeoutHeight := timeoutTx.GetTimeoutHeight()
	if timeoutHeight > 0 && uint64(sdkCtx.BlockHeight()) > timeoutHeight {
		return abci.ResponseDeliverTx{}, sdkerrors.Wrapf(
			sdkerrors.ErrTxTimeoutHeight, "block height: %d, timeout height: %d", sdkCtx.BlockHeight(), timeoutHeight,
		)
	}

	return txh.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (txh txTimeoutHeightMiddleware) SimulateTx(ctx context.Context, tx sdk.Tx, req txtypes.RequestSimulateTx) (txtypes.ResponseSimulateTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	timeoutTx, ok := tx.(TxWithTimeoutHeight)
	if !ok {
		return txtypes.ResponseSimulateTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "expected tx to implement TxWithTimeoutHeight")
	}

	timeoutHeight := timeoutTx.GetTimeoutHeight()
	if timeoutHeight > 0 && uint64(sdkCtx.BlockHeight()) > timeoutHeight {
		return txtypes.ResponseSimulateTx{}, sdkerrors.Wrapf(
			sdkerrors.ErrTxTimeoutHeight, "block height: %d, timeout height: %d", sdkCtx.BlockHeight(), timeoutHeight,
		)
	}

	return txh.next.SimulateTx(ctx, tx, req)
}

// validateMemoMiddleware will validate memo given the parameters passed in
// If memo is too large middleware returns with error, otherwise call next middleware
// CONTRACT: Tx must implement TxWithMemo interface
type validateMemoMiddleware struct {
	ak   AccountKeeper
	next txtypes.Handler
}

func ValidateMemoDecorator(ak AccountKeeper) txtypes.Middleware {
	return func(txHandler txtypes.Handler) txtypes.Handler {
		return validateMemoMiddleware{
			ak:   ak,
			next: txHandler,
		}
	}
}

var _ txtypes.Handler = indexEventsTxHandler{}

// CheckTx implements tx.Handler.CheckTx method.
func (vmd validateMemoMiddleware) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return abci.ResponseCheckTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	params := vmd.ak.GetParams(sdkCtx)

	memoLength := len(memoTx.GetMemo())
	if uint64(memoLength) > params.MaxMemoCharacters {
		return abci.ResponseCheckTx{}, sdkerrors.Wrapf(sdkerrors.ErrMemoTooLarge,
			"maximum number of characters is %d but received %d characters",
			params.MaxMemoCharacters, memoLength,
		)
	}

	return vmd.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (vmd validateMemoMiddleware) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return abci.ResponseDeliverTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	params := vmd.ak.GetParams(sdkCtx)

	memoLength := len(memoTx.GetMemo())
	if uint64(memoLength) > params.MaxMemoCharacters {
		return abci.ResponseDeliverTx{}, sdkerrors.Wrapf(sdkerrors.ErrMemoTooLarge,
			"maximum number of characters is %d but received %d characters",
			params.MaxMemoCharacters, memoLength,
		)
	}

	return vmd.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (vmd validateMemoMiddleware) SimulateTx(ctx context.Context, tx sdk.Tx, req txtypes.RequestSimulateTx) (txtypes.ResponseSimulateTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return txtypes.ResponseSimulateTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	params := vmd.ak.GetParams(sdkCtx)

	memoLength := len(memoTx.GetMemo())
	if uint64(memoLength) > params.MaxMemoCharacters {
		return txtypes.ResponseSimulateTx{}, sdkerrors.Wrapf(sdkerrors.ErrMemoTooLarge,
			"maximum number of characters is %d but received %d characters",
			params.MaxMemoCharacters, memoLength,
		)
	}

	return vmd.next.SimulateTx(ctx, tx, req)
}

var _ txtypes.Handler = consumeTxSizeGasMiddleware{}

// consumeTxSizeGasMiddleware will take in parameters and consume gas proportional
// to the size of tx before calling next AnteHandler. Note, the gas costs will be
// slightly over estimated due to the fact that any given signing account may need
// to be retrieved from state.
//
// CONTRACT: If simulate=true, then signatures must either be completely filled
// in or empty.
// CONTRACT: To use this decorator, signatures of transaction must be represented
// as legacytx.StdSignature otherwise simulate mode will incorrectly estimate gas cost.
type consumeTxSizeGasMiddleware struct {
	ak   AccountKeeper
	next txtypes.Handler
}

func ConsumeTxSizeGasMiddleware(ak AccountKeeper) txtypes.Middleware {
	return func(txHandler txtypes.Handler) txtypes.Handler {
		return consumeTxSizeGasMiddleware{
			ak:   ak,
			next: txHandler,
		}
	}
}

// CheckTx implements tx.Handler.CheckTx.
func (cgts consumeTxSizeGasMiddleware) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return abci.ResponseCheckTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	params := cgts.ak.GetParams(sdkCtx)
	sdkCtx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*sdk.Gas(len(sdkCtx.TxBytes())), "txSize")

	return cgts.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (cgts consumeTxSizeGasMiddleware) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return abci.ResponseDeliverTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	params := cgts.ak.GetParams(sdkCtx)
	sdkCtx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*sdk.Gas(len(sdkCtx.TxBytes())), "txSize")

	return cgts.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (cgts consumeTxSizeGasMiddleware) SimulateTx(ctx context.Context, tx sdk.Tx, req txtypes.RequestSimulateTx) (txtypes.ResponseSimulateTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return txtypes.ResponseSimulateTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	params := cgts.ak.GetParams(sdkCtx)
	sdkCtx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*sdk.Gas(len(sdkCtx.TxBytes())), "txSize")

	// simulate gas cost for signatures in simulate mode
	// in simulate mode, each element should be a nil signature
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return txtypes.ResponseSimulateTx{}, err
	}
	n := len(sigs)

	for i, signer := range sigTx.GetSigners() {
		// if signature is already filled in, no need to simulate gas cost
		if i < n && !isIncompleteSignature(sigs[i].Data) {
			continue
		}

		var pubkey cryptotypes.PubKey

		acc := cgts.ak.GetAccount(sdkCtx, signer)

		// use placeholder simSecp256k1Pubkey if sig is nil
		if acc == nil || acc.GetPubKey() == nil {
			pubkey = simSecp256k1Pubkey
		} else {
			pubkey = acc.GetPubKey()
		}

		// use stdsignature to mock the size of a full signature
		simSig := legacytx.StdSignature{ //nolint:staticcheck // this will be removed when proto is ready
			Signature: simSecp256k1Sig[:],
			PubKey:    pubkey,
		}

		sigBz := legacy.Cdc.MustMarshal(simSig)
		cost := sdk.Gas(len(sigBz) + 6)

		// If the pubkey is a multi-signature pubkey, then we estimate for the maximum
		// number of signers.
		if _, ok := pubkey.(*multisig.LegacyAminoPubKey); ok {
			cost *= params.TxSigLimit
		}

		sdkCtx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*cost, "txSize")
	}

	return cgts.next.SimulateTx(ctx, tx, req)
}

// isIncompleteSignature tests whether SignatureData is fully filled in for simulation purposes
func isIncompleteSignature(data signing.SignatureData) bool {
	if data == nil {
		return true
	}

	switch data := data.(type) {
	case *signing.SingleSignatureData:
		return len(data.Signature) == 0
	case *signing.MultiSignatureData:
		if len(data.Signatures) == 0 {
			return true
		}
		for _, s := range data.Signatures {
			if isIncompleteSignature(s) {
				return true
			}
		}
	}

	return false
}
