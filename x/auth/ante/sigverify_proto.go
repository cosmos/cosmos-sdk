package ante

import (
	"github.com/cosmos/cosmos-sdk/crypto/multisig"
	types2 "github.com/cosmos/cosmos-sdk/crypto/types"

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
		signerAcc, err := GetSignerAcc(ctx, svd.ak, signerAddrs[i])
		if err != nil {
			return ctx, err
		}

		signerAccs[i] = signerAcc

		signerInfo := sigTx.GetAuthInfo().SignerInfos[i]

		// retrieve pubkey
		pubKey := signerAccs[i].GetPubKey()
		if !simulate && pubKey == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
		}

		switch mi := signerInfo.ModeInfo.Sum.(type) {
		case *types.ModeInfo_Single_:
			single := mi.Single

			signBytes, err := svd.getSignBytesSingle(ctx, single, signerAcc, sigTx)
			if err != nil {
				return ctx, err
			}

			// verify signature
			if !simulate && !pubKey.VerifyBytes(signBytes, sig) {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "signature verification failed; verify correct account sequence and chain-id")
			}
		case *types.ModeInfo_Multi_:
			multisigPubKey, ok := pubKey.(multisig.MultisigPubKey)
			if !ok {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "key is not a multisig pubkey, but ModeInfo.Multi is used")
			}

			if !simulate {
				multiSigs, err := types2.DecodeMultisignatures(sig)
				if err != nil {
					return ctx, sdkerrors.Wrap(err, "cannot decode MultiSignature")
				}

				decodedMultisig := multisig.DecodedMultisignature{
					ModeInfo:   mi.Multi,
					Signatures: multiSigs,
				}

				if !multisigPubKey.VerifyMultisignature(
					func(single *types.ModeInfo_Single) ([]byte, error) {
						return svd.getSignBytesSingle(ctx, single, signerAcc, sigTx)
					}, decodedMultisig,
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

func (svd ProtoSigVerificationDecorator) getSignBytesSingle(ctx sdk.Context, single *types.ModeInfo_Single, signerAcc auth.AccountI, sigTx types.ProtoTx) ([]byte, error) {
	verifier := svd.handler
	genesis := ctx.BlockHeight() == 0
	var accNum uint64
	if !genesis {
		accNum = signerAcc.GetAccountNumber()
	}
	data := types.SigningData{
		ModeInfo:        single,
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
