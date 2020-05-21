package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

type ProtoSigVerificationDecorator struct {
	ak                AccountKeeper
	signModeVerifiers map[types.SignMode]signing.SignModeHandler
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
	sigTx, ok := tx.(signing.DecodedTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs := sigTx.GetSignatures()

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signerAddrs := sigTx.GetSigners()
	signerAccs := make([]auth.AccountI, len(signerAddrs))

	// check that signer length and signature length are the same
	if len(sigs) != len(signerAddrs) {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", len(signerAddrs), len(sigs))
	}

	for i, sig := range sigs {
		var signBytes []byte

		signerAcc, err := GetSignerAcc(ctx, svd.ak, signerAddrs[i])
		if err != nil {
			return ctx, err
		}

		signerAccs[i] = signerAcc

		signerInfo := sigTx.AuthInfo.SignerInfos[i]

		switch mi := signerInfo.ModeInfo.Sum.(type) {
		case *types.ModeInfo_Single_:
			single := mi.Single
			verifier, found := svd.signModeVerifiers[single.Mode]
			if !found {
				return ctx, fmt.Errorf("can't verify sign mode %s", single.Mode.String())
			}
			genesis := ctx.BlockHeight() == 0
			var accNum uint64
			if !genesis {
				accNum = signerAcc.GetAccountNumber()
			}
			data := signing.SigningData{
				ModeInfo:        single,
				PublicKey:       signerAcc.GetPubKey(),
				ChainID:         ctx.ChainID(),
				AccountNumber:   accNum,
				AccountSequence: signerAcc.GetSequence(),
			}
			signBytes, err = verifier.GetSignBytes(data, sigTx)
			if err != nil {
				return ctx, err
			}
		case *types.ModeInfo_Multi_:
			panic("TODO: can't handle multisignatures yet")
		default:
			panic("unexpected ModeInfo type")
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
	}

	return next(ctx, tx, simulate)
}
