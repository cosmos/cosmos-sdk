package middleware

import (
	"context"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// bankKeeper defines the contract needed for transferring tips.
type bankKeeper interface {
	SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error
}

type tipsTxHandler struct {
	next       tx.Handler
	bankKeeper bankKeeper
}

// TipsTxMiddleware TODO
func TipsTxMiddleware(txh tx.Handler) tx.Handler {
	return tipsTxHandler{next: txh}
}

var _ tx.Handler = tipsTxHandler{}

// CheckTx implements tx.Handler.CheckTx.
func (txh tipsTxHandler) CheckTx(ctx context.Context, sdkTx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	tipTx, err := assertTipTx(sdkTx)
	if err != nil {
		return txh.next.CheckTx(ctx, sdkTx, req)
	}

	if err := checkSigs(tipTx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	res, err := txh.next.CheckTx(ctx, sdkTx, req)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}

	if err := txh.transferTip(ctx, tipTx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return res, err
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh tipsTxHandler) DeliverTx(ctx context.Context, sdkTx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	tipTx, err := assertTipTx(sdkTx)
	if err != nil {
		return txh.next.DeliverTx(ctx, sdkTx, req)
	}

	if err := checkSigs(tipTx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	res, err := txh.next.DeliverTx(ctx, sdkTx, req)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	if err := txh.transferTip(ctx, tipTx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return res, err
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh tipsTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	tipTx, err := assertTipTx(sdkTx)
	if err != nil {
		return txh.next.SimulateTx(ctx, sdkTx, req)
	}

	if err := checkSigs(tipTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	res, err := txh.next.SimulateTx(ctx, sdkTx, req)
	if err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	if err := txh.transferTip(ctx, tipTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return res, err
}

func assertTipTx(sdkTx sdk.Tx) (tx.TipTx, error) {
	tipTx, ok := sdkTx.(tx.TipTx)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrTxDecode, "tx must be a TipTx, got %T", sdkTx)
	}

	return tipTx, nil
}

// checkSigs checks that all signatures in the tx have correct sign modes.
// Namely:
// - fee payer should sign over fees,
// - tipper should signer over tips.
func checkSigs(tipTx tx.TipTx) error {
	sigTx, ok := tipTx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrTxDecode, "tx must be a SigVerifiableTx, got %T", tipTx)
	}

	feePayer := tipTx.FeePayer()
	tipper := tipTx.GetTip().Tipper
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
	if err := checkCorrectSignModes(addr, sigData, func(signMode signing.SignMode) bool {
		return signMode == signing.SignMode_SIGN_MODE_DIRECT ||
			signMode == signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	}); err != nil {
		return sdkerrors.ErrUnauthorized.Wrapf("invalid sign mode for fee payer; %v", err)
	}

	return nil
}

// checkTipSigner checks whether a signature has signed over Tips.
func checkTipSigner(addr sdk.AccAddress, sigData signing.SignatureData) error {
	if err := checkCorrectSignModes(addr, sigData, func(signMode signing.SignMode) bool {
		return signMode == signing.SignMode_SIGN_MODE_DIRECT_AUX
	}); err != nil {
		return sdkerrors.ErrUnauthorized.Wrapf("invalid sign mode for tiper; %v", err)
	}

	return nil
}

// checkCorrectSignModes checks that all signatures are one of the allowed
// sign modes defined in `allowedSignModes`.
func checkCorrectSignModes(
	addr sdk.AccAddress,
	sigData signing.SignatureData,
	allowedSignModes func(signMode signing.SignMode) bool,
) error {
	switch sigData := sigData.(type) {
	case *signing.SingleSignatureData:
		{
			if !allowedSignModes(sigData.SignMode) {
				return fmt.Errorf("invalid sign mode for signer %s, got %s", addr, sigData.SignMode)
			}
		}
	case *signing.MultiSignatureData:
		{
			for _, s := range sigData.Signatures {
				err := checkCorrectSignModes(addr, s, allowedSignModes)
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

// transferTip transfers the tip from the tipper to the fee payer.
func (txh tipsTxHandler) transferTip(ctx context.Context, tipTx tx.TipTx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	tipper, err := sdk.AccAddressFromBech32(tipTx.GetTip().Tipper)
	if err != nil {
		return err
	}

	return txh.bankKeeper.SendCoins(sdkCtx, tipper, tipTx.FeePayer(), tipTx.GetTip().Amount)
}
