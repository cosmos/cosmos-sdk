package middleware

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type tipsTxHandler struct {
	next       tx.Handler
	bankKeeper types.BankKeeper
}

// NewTipsTxMiddleware returns a new middleware for handling meta-transactions
// with tips.
func NewTipsTxMiddleware(bankKeeper types.BankKeeper) tx.Middleware {
	return func(txh tx.Handler) tx.Handler {
		return tipsTxHandler{txh, bankKeeper}
	}
}

var _ tx.Handler = tipsTxHandler{}

// CheckTx implements tx.Handler.CheckTx.
func (txh tipsTxHandler) CheckTx(ctx context.Context, sdkTx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	res, err := txh.next.CheckTx(ctx, sdkTx, req)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}

	tipTx, ok := sdkTx.(tx.TipTx)
	if !ok || tipTx.GetTip() == nil {
		return res, err
	}

	if err := txh.transferTip(ctx, tipTx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return res, err
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh tipsTxHandler) DeliverTx(ctx context.Context, sdkTx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	res, err := txh.next.DeliverTx(ctx, sdkTx, req)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	tipTx, ok := sdkTx.(tx.TipTx)
	if !ok || tipTx.GetTip() == nil {
		return res, err
	}

	if err := txh.transferTip(ctx, tipTx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return res, err
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh tipsTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	res, err := txh.next.SimulateTx(ctx, sdkTx, req)
	if err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	tipTx, ok := sdkTx.(tx.TipTx)
	if !ok || tipTx.GetTip() == nil {
		return res, err
	}

	if err := txh.transferTip(ctx, tipTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return res, err
}

// transferTip transfers the tip from the tipper to the fee payer.
func (txh tipsTxHandler) transferTip(ctx context.Context, tipTx tx.TipTx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	tipper, err := sdk.AccAddressFromBech32(tipTx.GetTip().Tipper)
	if err != nil {
		return err
	}

	return txh.bankKeeper.SendCoins(sdkCtx, tipper, tipTx.FeePayer(), tipTx.GetTip().Amount)
}

type signModeTxHandler struct {
	next tx.Handler
}

// SignModeTxMiddleware returns a new middleware that checks that
// all signatures in the tx have correct sign modes.
// Namely:
// - fee payer should sign over fees,
// - tipper should signer over tips.
func SignModeTxMiddleware(txh tx.Handler) tx.Handler {
	return signModeTxHandler{txh}
}

var _ tx.Handler = signModeTxHandler{}

// CheckTx implements tx.Handler.CheckTx.
func (txh signModeTxHandler) CheckTx(ctx context.Context, sdkTx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := checkSignMode(sdkTx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return txh.next.CheckTx(ctx, sdkTx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh signModeTxHandler) DeliverTx(ctx context.Context, sdkTx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := checkSignMode(sdkTx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return txh.next.DeliverTx(ctx, sdkTx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh signModeTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := checkSignMode(sdkTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return txh.next.SimulateTx(ctx, sdkTx, req)
}

// checkSignMode checks that all signatures in the tx have correct sign modes.
// Namely:
// - fee payer should sign over fees,
// - tipper should signer over tips.
func checkSignMode(sdkTx sdk.Tx) error {
	tipTx, ok := sdkTx.(tx.TipTx)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrTxDecode, "tx must be a TipTx, got %T", sdkTx)
	}
	feePayer := tipTx.FeePayer()
	var tipper string
	if tip := tipTx.GetTip(); tip != nil {
		tipper = tipTx.GetTip().Tipper
	}

	sigTx, ok := sdkTx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrTxDecode, "tx must be a SigVerifiableTx, got %T", sdkTx)
	}
	sigsV2, err := sigTx.GetSignaturesV2()
	if err != nil {
		return err
	}

	for _, sig := range sigsV2 {
		addr := sdk.AccAddress(sig.PubKey.Address())

		// Make sure the feePayer signs over the Fee.
		if addr.Equals(feePayer) {
			if err := checkFeeSigner(addr, sig.Data); err != nil {
				return err
			}
		}

		// Make sure the tipper signs over the Tip.
		if addr.String() == tipper {
			if err := checkTipSigner(addr, sig.Data); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkFeeSigner checks whether a signature has signed over Fees.
func checkFeeSigner(addr sdk.AccAddress, sigData signing.SignatureData) error {
	if err := checkCorrectSignModes(addr, sigData, []signing.SignMode{
		signing.SignMode_SIGN_MODE_DIRECT,
		signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}, "fee payer"); err != nil {
		return err
	}

	return nil
}

// checkTipSigner checks whether a signature has signed over Tips.
func checkTipSigner(addr sdk.AccAddress, sigData signing.SignatureData) error {
	if err := checkCorrectSignModes(addr, sigData, []signing.SignMode{
		signing.SignMode_SIGN_MODE_DIRECT,
		signing.SignMode_SIGN_MODE_DIRECT_AUX,
		signing.SignMode_SIGN_MODE_AMINO_AUX,
	}, "tipper"); err != nil {
		return err
	}

	return nil
}

// checkCorrectSignModes checks that all signatures are one of the allowed
// sign modes defined in `allowedSignModes`.
// For multi-signatures, check recursively.
func checkCorrectSignModes(
	addr sdk.AccAddress,
	sigData signing.SignatureData,
	allowedSignModes []signing.SignMode,
	errString string,
) error {
	switch sigData := sigData.(type) {
	case *signing.SingleSignatureData:
		{
			if !arrayIncludes(allowedSignModes, sigData.SignMode) {
				return sdkerrors.ErrUnauthorized.Wrapf("invalid sign mode for %s %s, got %s", errString, addr, sigData.SignMode)
			}
		}
	case *signing.MultiSignatureData:
		{
			for _, s := range sigData.Signatures {
				err := checkCorrectSignModes(addr, s, allowedSignModes, errString)
				if err != nil {
					return err
				}
			}
		}
	default:
		return sdkerrors.ErrInvalidType.Wrapf("got unexpected SignatureData %T", sigData)
	}

	return nil
}

func arrayIncludes(sms []signing.SignMode, sm signing.SignMode) bool {
	for _, x := range sms {
		if x == sm {
			return true
		}
	}

	return false
}
