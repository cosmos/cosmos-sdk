package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/raw"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

type SignModeVerifier interface {
	GetSignBytes(tx raw.DecodedTx) ([]byte, error)
}

type ProtoSigVerificationDecorator struct {
	ak                AccountKeeper
	signModeVerifiers map[types.SignMode]SignModeVerifier
}

func NewProtoSigVerificationDecorator(ak AccountKeeper) ProtoSigVerificationDecorator {
	return ProtoSigVerificationDecorator{
		ak: ak,
	}
}

func (svd ProtoSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// no need to verify signatures on recheck tx
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}
	sigTx, ok := tx.(raw.DecodedTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs := sigTx.GetSignatures()

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signerAddrs := sigTx.GetSigners()
	signerAccs := make([]exported.Account, len(signerAddrs))

	for _, si := range sigTx.AuthInfo.SignerInfos {
		switch mi := si.ModeInfo.Sum.(type) {
		case *types.ModeInfo_Single_:
			verifier, found := svd.signModeVerifiers[mi.Single.Mode]
			if !found {
				return ctx, fmt.Errorf("can't verify sign mode %s", mi.Single.Mode.String())
			}
			bz, err := verifier.GetSignBytes(sigTx)
			if err != nil {
				return ctx, err
			}
			// retrieve pubkey
			pubKey := signerAccs[i].GetPubKey()
			if !simulate && pubKey == nil {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
			}

			// verify signature
			if !simulate && !pubKey.VerifyBytes(signBytes, sig) {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "signature verification failed; verify correct account sequence and chain-id")
			}
		case *types.ModeInfo_Multi_:
		}
	}

	// check that signer length and signature length are the same
	if len(sigs) != len(signerAddrs) {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", len(signerAddrs), len(sigs))
	}

	for i, sig := range sigs {
		signerAccs[i], err = GetSignerAcc(ctx, svd.ak, signerAddrs[i])
		if err != nil {
			return ctx, err
		}

		// retrieve signBytes of tx
		signBytes := sigTx.GetSignBytes(ctx, signerAccs[i])

		// retrieve pubkey
		pubKey := signerAccs[i].GetPubKey()
		if !simulate && pubKey == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
		}

		// verify signature
		if !simulate && !pubKey.VerifyBytes(signBytes, sig) {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "signature verification failed; verify correct account sequence and chain-id")
		}
	}

	return next(ctx, tx, simulate)
}
