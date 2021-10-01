package middleware

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	abci "github.com/tendermint/tendermint/abci/types"
)

type validateBasicTxHandler struct {
	next tx.Handler
}

// ValidateBasicMiddleware will call tx.ValidateBasic, msg.ValidateBasic(for each msg inside tx)
// and return any non-nil error.
// If ValidateBasic passes, middleware calls next middleware in chain. Note,
// validateBasicTxHandler will not get executed on ReCheckTx since it
// is not dependent on application state.
func ValidateBasicMiddleware(txh tx.Handler) tx.Handler {
	return validateBasicTxHandler{
		next: txh,
	}
}

var _ tx.Handler = validateBasicTxHandler{}

// validateBasicTxMsgs executes basic validator calls for messages.
func validateBasicTxMsgs(msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "must contain at least one message")
	}

	for _, msg := range msgs {
		err := msg.ValidateBasic()
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (txh validateBasicTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	// no need to validate basic on recheck tx, call next middleware
	if req.Type == abci.CheckTxType_Recheck {
		return txh.next.CheckTx(ctx, tx, req)
	}

	if err := validateBasicTxMsgs(tx.GetMsgs()); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	if err := tx.ValidateBasic(); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return txh.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh validateBasicTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := tx.ValidateBasic(); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	if err := validateBasicTxMsgs(tx.GetMsgs()); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return txh.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (txh validateBasicTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := sdkTx.ValidateBasic(); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	if err := validateBasicTxMsgs(sdkTx.GetMsgs()); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return txh.next.SimulateTx(ctx, sdkTx, req)
}

var _ tx.Handler = txTimeoutHeightTxHandler{}

type txTimeoutHeightTxHandler struct {
	next tx.Handler
}

// TxTimeoutHeightMiddleware defines a middleware that checks for a
// tx height timeout.
func TxTimeoutHeightMiddleware(txh tx.Handler) tx.Handler {
	return txTimeoutHeightTxHandler{
		next: txh,
	}
}

func checkTimeout(ctx context.Context, tx sdk.Tx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	timeoutTx, ok := tx.(sdk.TxWithTimeoutHeight)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "expected tx to implement TxWithTimeoutHeight")
	}

	timeoutHeight := timeoutTx.GetTimeoutHeight()
	if timeoutHeight > 0 && uint64(sdkCtx.BlockHeight()) > timeoutHeight {
		return sdkerrors.Wrapf(
			sdkerrors.ErrTxTimeoutHeight, "block height: %d, timeout height: %d", sdkCtx.BlockHeight(), timeoutHeight,
		)
	}

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (txh txTimeoutHeightTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := checkTimeout(ctx, tx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return txh.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh txTimeoutHeightTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := checkTimeout(ctx, tx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return txh.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (txh txTimeoutHeightTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := checkTimeout(ctx, sdkTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return txh.next.SimulateTx(ctx, sdkTx, req)
}

type validateMemoTxHandler struct {
	ak   AccountKeeper
	next tx.Handler
}

// ValidateMemoMiddleware will validate memo given the parameters passed in
// If memo is too large middleware returns with error, otherwise call next middleware
// CONTRACT: Tx must implement TxWithMemo interface
func ValidateMemoMiddleware(ak AccountKeeper) tx.Middleware {
	return func(txHandler tx.Handler) tx.Handler {
		return validateMemoTxHandler{
			ak:   ak,
			next: txHandler,
		}
	}
}

var _ tx.Handler = validateMemoTxHandler{}

func (vmm validateMemoTxHandler) checkForValidMemo(ctx context.Context, tx sdk.Tx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	params := vmm.ak.GetParams(sdkCtx)

	memoLength := len(memoTx.GetMemo())
	if uint64(memoLength) > params.MaxMemoCharacters {
		return sdkerrors.Wrapf(sdkerrors.ErrMemoTooLarge,
			"maximum number of characters is %d but received %d characters",
			params.MaxMemoCharacters, memoLength,
		)
	}

	return nil
}

// CheckTx implements tx.Handler.CheckTx method.
func (vmm validateMemoTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := vmm.checkForValidMemo(ctx, tx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return vmm.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (vmm validateMemoTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := vmm.checkForValidMemo(ctx, tx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return vmm.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (vmm validateMemoTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := vmm.checkForValidMemo(ctx, sdkTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return vmm.next.SimulateTx(ctx, sdkTx, req)
}

var _ tx.Handler = consumeTxSizeGasTxHandler{}

type consumeTxSizeGasTxHandler struct {
	ak   AccountKeeper
	next tx.Handler
}

// ConsumeTxSizeGasMiddleware will take in parameters and consume gas proportional
// to the size of tx before calling next middleware. Note, the gas costs will be
// slightly over estimated due to the fact that any given signing account may need
// to be retrieved from state.
//
// CONTRACT: If simulate=true, then signatures must either be completely filled
// in or empty.
// CONTRACT: To use this middleware, signatures of transaction must be represented
// as legacytx.StdSignature otherwise simulate mode will incorrectly estimate gas cost.
func ConsumeTxSizeGasMiddleware(ak AccountKeeper) tx.Middleware {
	return func(txHandler tx.Handler) tx.Handler {
		return consumeTxSizeGasTxHandler{
			ak:   ak,
			next: txHandler,
		}
	}
}

func (cgts consumeTxSizeGasTxHandler) simulateSigGasCost(ctx context.Context, tx sdk.Tx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := cgts.ak.GetParams(sdkCtx)

	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}

	// in simulate mode, each element should be a nil signature
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return err
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

	return nil
}

func (cgts consumeTxSizeGasTxHandler) consumeTxSizeGas(ctx context.Context, tx sdk.Tx, txBytes []byte, simulate bool) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := cgts.ak.GetParams(sdkCtx)
	sdkCtx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*sdk.Gas(len(txBytes)), "txSize")

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (cgts consumeTxSizeGasTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := cgts.consumeTxSizeGas(ctx, tx, req.GetTx(), false); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return cgts.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (cgts consumeTxSizeGasTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := cgts.consumeTxSizeGas(ctx, tx, req.GetTx(), false); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return cgts.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (cgts consumeTxSizeGasTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := cgts.consumeTxSizeGas(ctx, sdkTx, req.TxBytes, true); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	if err := cgts.simulateSigGasCost(ctx, sdkTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return cgts.next.SimulateTx(ctx, sdkTx, req)
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
