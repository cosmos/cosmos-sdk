package ante

import (
	"github.com/cosmos/cosmos-sdk/crypto/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type ProtoSigVerificationDecorator struct {
	ak      AccountKeeper
	handler types.SignModeHandler
}

func NewProtoSigVerificationDecorator(ak AccountKeeper, signModeHandler types.SignModeHandler) ProtoSigVerificationDecorator {
	return ProtoSigVerificationDecorator{
		ak:      ak,
		handler: signModeHandler,
	}
}

func (svd ProtoSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// no need to verify signatures on recheck tx
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}
	sigTx, ok := tx.(types.ProtoTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs, err := sigTx.GetSignatureData()
	if err != nil {
		return ctx, err
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signerAddrs := sigTx.GetSigners()
	signerAccs := make([]auth.AccountI, len(signerAddrs))

	// check that signer length and signature length are the same
	if len(sigs) != len(signerAddrs) {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", len(signerAddrs), len(sigs))
	}

	for i, sig := range sigs {
		signerAcc, err := GetSignerAcc(ctx, svd.ak, signerAddrs[i])
		if err != nil {
			return ctx, err
		}

		signerAccs[i] = signerAcc

		// retrieve pubkey
		pubKey := signerAccs[i].GetPubKey()
		if !simulate && pubKey == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
		}

		switch sig := sig.(type) {
		case *types.SingleSignatureData:
			signBytes, err := svd.getSignBytesSingle(ctx, sig.SignMode, signerAcc, sigTx)
			if err != nil {
				return ctx, err
			}

			// verify signature
			if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "signature verification failed; verify correct account sequence and chain-id")
			}
		case *types.MultiSignatureData:
			multisigPubKey, ok := pubKey.(multisig.MultisigPubKey)
			if !ok {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "key is not a multisig pubkey, but ModeInfo.Multi is used")
			}

			if !simulate {
				if !multisigPubKey.VerifyMultisignature(
					func(mode types.SignMode) ([]byte, error) {
						return svd.getSignBytesSingle(ctx, mode, signerAcc, sigTx)
					}, sig,
				) {
					return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "signature verification failed; verify correct account sequence and chain-id")
				}
			}
		default:
			panic("unexpected ModeInfo type")
		}
	}

	return next(ctx, tx, simulate)
}

func (svd ProtoSigVerificationDecorator) getSignBytesSingle(ctx sdk.Context, signMode types.SignMode, signerAcc auth.AccountI, sigTx types.ProtoTx) ([]byte, error) {
	verifier := svd.handler
	genesis := ctx.BlockHeight() == 0
	var accNum uint64
	if !genesis {
		accNum = signerAcc.GetAccountNumber()
	}
	data := types.SigningData{
		Mode:            signMode,
		PublicKey:       signerAcc.GetPubKey(),
		ChainID:         ctx.ChainID(),
		AccountNumber:   accNum,
		AccountSequence: signerAcc.GetSequence(),
	}
	signBytes, err := verifier.GetSignBytes(data, sigTx)
	if err != nil {
		return nil, err
	}
	return signBytes, nil
}
