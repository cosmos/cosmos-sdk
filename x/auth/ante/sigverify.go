package ante

import (
	"bytes"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errs "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var (
	// simulation signature values used to estimate gas consumption
	simSecp256k1Pubkey secp256k1.PubKeySecp256k1
	simSecp256k1Sig    [64]byte
)

func init() {
	// This decodes a valid hex string into a sepc256k1Pubkey for use in transaction simulation
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(simSecp256k1Pubkey[:], bz)
}

// SignatureVerificationGasConsumer is the type of function that is used to both consume gas when verifying signatures
// and also to accept or reject different types of PubKey's. This is where apps can define their own PubKey
type SignatureVerificationGasConsumer = func(meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params types.Params) error

// Consume parameter-defined amount of gas for each signature according to the passed-in SignatureVerificationGasConsumer function
// before calling the next AnteHandler
type SigGasConsumeDecorator struct {
	ak             keeper.AccountKeeper
	sigGasConsumer SignatureVerificationGasConsumer
}

func NewSigGasConsumeDecorator(ak keeper.AccountKeeper, sigGasConsumer SignatureVerificationGasConsumer) SigGasConsumeDecorator {
	return SigGasConsumeDecorator{
		ak:             ak,
		sigGasConsumer: sigGasConsumer,
	}
}

func (sgcd SigGasConsumeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	stdTx, ok := tx.(types.StdTx)
	if !ok {
		return ctx, errs.Wrap(errs.ErrTxDecode, "Tx must be a StdTx")
	}

	params := sgcd.ak.GetParams(ctx)

	stdSigs := stdTx.GetSignatures()

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signerAddrs := stdTx.GetSigners()

	for i, sig := range stdSigs {
		pubKey := sig.PubKey
		if pubKey == nil {
			signerAcc, err := GetSignerAcc(ctx, sgcd.ak, signerAddrs[i])
			if err != nil {
				return ctx, err
			}
			pubKey = signerAcc.GetPubKey()
		}

		if simulate {
			// Simulated txs should not contain a signature and are not required to
			// contain a pubkey, so we must account for tx size of including a
			// StdSignature (Amino encoding) and simulate gas consumption
			// (assuming a SECP256k1 simulation key).
			consumeSimSigGas(ctx.GasMeter(), pubKey, sig, params)
		}
		err = sgcd.sigGasConsumer(ctx.GasMeter(), sig.Signature, pubKey, params)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// Verify all signatures for tx and return error if any are invalid
// increment sequence of each signer and set updated account back in store
// Call next AnteHandler
type SigVerificationDecorator struct {
	ak keeper.AccountKeeper
}

func NewSigVerificationDecorator(ak keeper.AccountKeeper) SigVerificationDecorator {
	return SigVerificationDecorator{
		ak: ak,
	}
}

func (svd SigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	stdTx, ok := tx.(types.StdTx)
	if !ok {
		return ctx, errs.Wrap(errs.ErrTxDecode, "Tx must be a StdTx")
	}

	isGenesis := ctx.BlockHeight() == 0

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	stdSigs := stdTx.GetSignatures()

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signerAddrs := stdTx.GetSigners()
	signerAccs := make([]exported.Account, len(signerAddrs))

	// check that signer length and signature length are the same
	if len(stdSigs) != len(signerAddrs) {
		return ctx, errs.Wrapf(errs.ErrUnauthorized, "Wrong number of signers. Expected: %d, got %d", len(signerAddrs), len(stdSigs))
	}

	for i := 0; i < len(stdSigs); i++ {
		signerAccs[i], err = GetSignerAcc(ctx, svd.ak, signerAddrs[i])
		if err != nil {
			return ctx, err
		}

		// check signature, return account with incremented nonce
		signBytes := GetSignBytes(ctx.ChainID(), stdTx, signerAccs[i], isGenesis)
		signerAccs[i], err = processSig(ctx, signerAccs[i], stdSigs[i], signBytes, simulate)
		if err != nil {
			return ctx, err
		}

		svd.ak.SetAccount(ctx, signerAccs[i])
	}

	return next(ctx, tx, simulate)
}

// ValidateSigCountDecorator takes in Params and returns errors if there are too many signatures in the tx for the given params
// otherwise it calls next AnteHandler
type ValidateSigCountDecorator struct {
	ak keeper.AccountKeeper
}

func NewValidateSigCountDecorator(ak keeper.AccountKeeper) ValidateSigCountDecorator {
	return ValidateSigCountDecorator{
		ak: ak,
	}
}

func (vscd ValidateSigCountDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	stdTx, ok := tx.(types.StdTx)
	if !ok {
		return ctx, errs.Wrap(errs.ErrTxDecode, "Tx must be a StdTx")
	}

	params := vscd.ak.GetParams(ctx)

	stdSigs := stdTx.GetSignatures()

	sigCount := 0
	for i := 0; i < len(stdSigs); i++ {
		sigCount += types.CountSubKeys(stdSigs[i].PubKey)
		if uint64(sigCount) > params.TxSigLimit {
			return ctx, errs.Wrapf(errs.ErrTooManySignatures,
				"signatures: %d, limit: %d", sigCount, params.TxSigLimit)
		}
	}

	return next(ctx, tx, simulate)
}

type SetPubKeyDecorator struct {
	ak keeper.AccountKeeper
}

func NewSetPubKeyDecorator(ak keeper.AccountKeeper) SetPubKeyDecorator {
	return SetPubKeyDecorator{
		ak: ak,
	}
}

func (spkd SetPubKeyDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if simulate {
		return next(ctx, tx, simulate)
	}
	stdTx, ok := tx.(types.StdTx)
	if !ok {
		return ctx, errs.Wrap(errs.ErrTxDecode, "Tx must be a StdTx")
	}

	stdSigs := stdTx.GetSignatures()
	signers := stdTx.GetSigners()

	for i, sig := range stdSigs {
		if sig.PubKey != nil {
			if !bytes.Equal(sig.PubKey.Address(), signers[i]) {
				return ctx, errs.Wrapf(errs.ErrInvalidPubKey,
					"PubKey does not match Signer address %s with signer index: %d", signers[i], i)
			}

			acc, err := GetSignerAcc(ctx, spkd.ak, signers[i])
			if err != nil {
				return ctx, err
			}
			err = acc.SetPubKey(sig.PubKey)
			if err != nil {
				return ctx, errs.Wrap(errs.ErrInvalidPubKey, err.Error())
			}
			spkd.ak.SetAccount(ctx, acc)
		}
	}

	return next(ctx, tx, simulate)
}

// DefaultSigVerificationGasConsumer is the default implementation of SignatureVerificationGasConsumer. It consumes gas
// for signature verification based upon the public key type. The cost is fetched from the given params and is matched
// by the concrete type.
func DefaultSigVerificationGasConsumer(
	meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params types.Params,
) error {
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

func consumeSimSigGas(gasmeter sdk.GasMeter, pubkey crypto.PubKey, sig types.StdSignature, params types.Params) {
	simSig := types.StdSignature{PubKey: pubkey}
	if len(sig.Signature) == 0 {
		simSig.Signature = simSecp256k1Sig[:]
	}

	sigBz := types.ModuleCdc.MustMarshalBinaryLengthPrefixed(simSig)
	cost := sdk.Gas(len(sigBz) + 6)

	// If the pubkey is a multi-signature pubkey, then we estimate for the maximum
	// number of signers.
	if _, ok := pubkey.(multisig.PubKeyMultisigThreshold); ok {
		cost *= params.TxSigLimit
	}

	gasmeter.ConsumeGas(params.TxSizeCostPerByte*cost, "txSize")
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
			return simSecp256k1Pubkey, nil
		}

		return pubKey, nil
	}

	if pubKey == nil {
		pubKey = sig.PubKey
		if pubKey == nil {
			return nil, errs.Wrap(errs.ErrInvalidPubKey, "PubKey not found")
		}

		if !bytes.Equal(pubKey.Address(), acc.GetAddress()) {
			return nil, errs.Wrapf(errs.ErrUnauthorized,
				"PubKey does not match Signer address %s", acc.GetAddress())
		}
	}

	return pubKey, nil
}

// verify the signature and increment the sequence. If the account doesn't have
// a pubkey, set it.
func processSig(
	ctx sdk.Context, acc exported.Account, sig types.StdSignature, signBytes []byte, simulate bool,
) (updatedAcc exported.Account, err error) {

	pubKey := acc.GetPubKey()
	if !simulate && pubKey == nil {
		return nil, errs.Wrap(errs.ErrInvalidPubKey, "pubkey on account is not set")
	}

	if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, errs.Wrap(errs.ErrUnauthorized, "signature verification failed; verify correct account sequence and chain-id")
	}

	if err = acc.SetSequence(acc.GetSequence() + 1); err != nil {
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

// GetSignBytes returns a slice of bytes to sign over for a given transaction
// and an account.
func GetSignBytes(chainID string, stdTx types.StdTx, acc exported.Account, genesis bool) []byte {
	var accNum uint64
	if !genesis {
		accNum = acc.GetAccountNumber()
	}

	return types.StdSignBytes(
		chainID, accNum, acc.GetSequence(), stdTx.Fee, stdTx.Msgs, stdTx.Memo,
	)
}
