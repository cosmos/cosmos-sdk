package ante

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// SignatureVerificationGasConsumer is the type of function that is used to both consume gas when verifying signatures
// and also to accept or reject different types of PubKey's. This is where apps can define their own PubKey
type SignatureVerificationGasConsumer = func(meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params types.Params) error

func SigGasConsumeDecorator(ak keeper.AccountKeeper, sigGasConsumer SignatureVerificationGasConsumer) {
	return func(ctx Context, tx Tx, simulate bool, next AnteHandler) {
		stdTx, ok := tx.(types.StdTx)
		if !ok {
			return ctx, errs.Wrap(errs.ErrInternal, "Tx must be a StdTx")
		}

		stdSigs := stdTx.GetSignatures()
		params := ak.GetParams(ctx)

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		signerAddrs := stdTx.GetSigners()
		signerAccs := make([]exported.Account, len(signerAddrs))

		for i, sig := range stdSigs {
			pubKey := signerAccs[i].GetPubKey()
			if simulate {
				// Simulated txs should not contain a signature and are not required to
				// contain a pubkey, so we must account for tx size of including a
				// StdSignature (Amino encoding) and simulate gas consumption
				// (assuming a SECP256k1 simulation key).
				consumeSimSigGas(ctx.GasMeter(), pubKey, sig, params)
			} else {
				err := sigGasConsumer(ctx.GasMeter(), sig.Signature, pubKey, params)
				if err != nil {
					return ctx, err
				}
			}
		}

		return next(ctx, tx, simulate)
	}
}

func SigVerificationDecorator(ak keeper.AccountKeeper) {
	return func(ctx Context, tx Tx, simulate bool, next sdk.AnteHandler) (newCtx, error) {
		stdTx, ok := tx.(types.StdTx)
		if !ok {
			return ctx, errs.Wrap(errs.ErrInternal, "Tx must be a StdTx")
		}

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		stdSigs := stdTx.GetSignatures()

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		signerAddrs := stdTx.GetSigners()
		signerAccs := make([]exported.Account, len(signerAddrs))

		for i := 0; i < len(stdSigs); i++ {
			signerAccs[i], err = GetSignerAcc(ctx, ak, signerAddrs[i])
			if err != nil {
				return ctx, err
			}

			// check signature, return account with incremented nonce
			signBytes := GetSignBytes(newCtx.ChainID(), stdTx, signerAccs[i], isGenesis)
			signerAccs[i], err = processSig(newCtx, signerAccs[i], stdSigs[i], signBytes, simulate)
			if err != nil {
				return ctx, err
			}

			ak.SetAccount(ctx, signerAccs[i])
		}

		return next(ctx, tx, simulate)
	}
}

// DefaultSigVerificationGasConsumer is the default implementation of SignatureVerificationGasConsumer. It consumes gas
// for signature verification based upon the public key type. The cost is fetched from the given params and is matched
// by the concrete type.
func DefaultSigVerificationGasConsumer(
	meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params types.Params,
) sdk.Result {
	switch pubkey := pubkey.(type) {
	case ed25519.PubKeyEd25519:
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return errs.Wrap(errs.ErrInvalidPubKey, "ED25519 public keys are unsupported")

	case secp256k1.PubKeySecp256k1:
		meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: secp256k1")
		return nil

	case multisig.PubKeyMultisigThreshold:
		var multisignature multisig.Multisignature
		codec.Cdc.MustUnmarshalBinaryBare(sig, &multisignature)

		consumeMultisignatureVerificationGas(meter, multisignature, pubkey, params)
		return nil

	default:
		return errs.Wrapf(errs.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
	}
}

func consumeMultisignatureVerificationGas(meter sdk.GasMeter,
	sig multisig.Multisignature, pubkey multisig.PubKeyMultisigThreshold,
	params types.Params) {

	size := sig.BitArray.Size()
	sigIndex := 0
	for i := 0; i < size; i++ {
		if sig.BitArray.GetIndex(i) {
			DefaultSigVerificationGasConsumer(meter, sig.Sigs[sigIndex], pubkey.PubKeys[i], params)
			sigIndex++
		}
	}
}

// ProcessPubKey verifies that the given account address matches that of the
// StdSignature. In addition, it will set the public key of the account if it
// has not been set.
func ProcessPubKey(acc exported.Account, sig types.StdSignature, simulate bool) (crypto.PubKey, error) {
	// If pubkey is not known for account, set it from the types.StdSignature.
	pubKey := acc.GetPubKey()
	if simulate {
		// In simulate mode the transaction comes with no signatures, thus if the
		// account's pubkey is nil, both signature verification and gasKVStore.Set()
		// shall consume the largest amount, i.e. it takes more gas to verify
		// secp256k1 keys than ed25519 ones.
		if pubKey == nil {
			return simSecp256k1Pubkey, sdk.Result{}
		}

		return pubKey, nil
	}

	if pubKey == nil {
		pubKey = sig.PubKey
		if pubKey == nil {
			return nil, errs.Wrap(errs.ErrInvalidPubKey, "PubKey not found")
		}

		if !bytes.Equal(pubKey.Address(), acc.GetAddress()) {
			return nil, errs.Wrapf(errs.ErrInvalidPubKey,
				"PubKey does not match Signer address %s", acc.GetAddress())
		}
	}

	return pubKey, nil
}

// verify the signature and increment the sequence. If the account doesn't have
// a pubkey, set it.
func processSig(
	ctx sdk.Context, acc exported.Account, sig types.StdSignature, signBytes []byte, simulate bool,
) (updatedAcc exported.Account, res sdk.Result) {

	pubKey, err := ProcessPubKey(acc, sig, simulate)
	if err != nil {
		return nil, res
	}

	err := acc.SetPubKey(pubKey)
	if err != nil {
		return nil, errs.Wrap(errs.ErrInternal, "setting PubKey on signer's account")
	}

	if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, errs.Wrap(errs.ErrUnauthorized, "signature verification failed; verify correct account sequence and chain-id")
	}

	if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
		panic(err)
	}

	return acc, nil
}

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak keeper.AccountKeeper, addr sdk.AccAddress) (exported.Account, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}
	return nil, errs.Wrapf(errs.ErrUnknownAddress, "account %s does not exist", addr)
}
