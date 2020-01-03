package ante

import (
	"bytes"
	"encoding/hex"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	// simulation signature values used to estimate gas consumption
	simSecp256k1Pubkey secp256k1.PubKeySecp256k1
	simSecp256k1Sig    [64]byte

	_ SigVerifiableTx = (*types.StdTx)(nil) // assert StdTx implements SigVerifiableTx
)

func init() {
	// This decodes a valid hex string into a sepc256k1Pubkey for use in transaction simulation
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(simSecp256k1Pubkey[:], bz)
}

// SignatureVerificationGasConsumer is the type of function that is used to both
// consume gas when verifying signatures and also to accept or reject different types of pubkeys
// This is where apps can define their own PubKey
type SignatureVerificationGasConsumer = func(meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params types.Params) error

// SigVerifiableTx defines a Tx interface for all signature verification decorators
type SigVerifiableTx interface {
	sdk.Tx
	GetSignatures() [][]byte
	GetSigners() []sdk.AccAddress
	GetPubKeys() []crypto.PubKey // If signer already has pubkey in context, this list will have nil in its place
	GetSignBytes(ctx sdk.Context, acc exported.Account) []byte
}

// SetPubKeyDecorator sets PubKeys in context for any signer which does not already have pubkey set
// PubKeys must be set in context for all signers before any other sigverify decorators run
// CONTRACT: Tx must implement SigVerifiableTx interface
type SetPubKeyDecorator struct {
	ak keeper.AccountKeeper
}

func NewSetPubKeyDecorator(ak keeper.AccountKeeper) SetPubKeyDecorator {
	return SetPubKeyDecorator{
		ak: ak,
	}
}

func (spkd SetPubKeyDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}

	pubkeys := sigTx.GetPubKeys()
	signers := sigTx.GetSigners()

	for i, pk := range pubkeys {
		// PublicKey was omitted from slice since it has already been set in context
		if pk == nil {
			if !simulate {
				continue
			}
			pk = simSecp256k1Pubkey
		}
		// Only make check if simulate=false
		if !simulate && !bytes.Equal(pk.Address(), signers[i]) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey,
				"pubKey does not match signer address %s with signer index: %d", signers[i], i)
		}

		acc, err := GetSignerAcc(ctx, spkd.ak, signers[i])
		if err != nil {
			return ctx, err
		}
		// account already has pubkey set,no need to reset
		if acc.GetPubKey() != nil {
			continue
		}
		err = acc.SetPubKey(pk)
		if err != nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, err.Error())
		}
		spkd.ak.SetAccount(ctx, acc)
	}

	return next(ctx, tx, simulate)
}

// Consume parameter-defined amount of gas for each signature according to the passed-in SignatureVerificationGasConsumer function
// before calling the next AnteHandler
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
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
	sigTx, ok := tx.(SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	params := sgcd.ak.GetParams(ctx)
	sigs := sigTx.GetSignatures()

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signerAddrs := sigTx.GetSigners()

	for i, sig := range sigs {
		signerAcc, err := GetSignerAcc(ctx, sgcd.ak, signerAddrs[i])
		if err != nil {
			return ctx, err
		}
		pubKey := signerAcc.GetPubKey()

		if simulate && pubKey == nil {
			// In simulate mode the transaction comes with no signatures, thus if the
			// account's pubkey is nil, both signature verification and gasKVStore.Set()
			// shall consume the largest amount, i.e. it takes more gas to verify
			// secp256k1 keys than ed25519 ones.
			if pubKey == nil {
				pubKey = simSecp256k1Pubkey
			}
		}
		err = sgcd.sigGasConsumer(ctx.GasMeter(), sig, pubKey, params)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// Verify all signatures for a tx and return an error if any are invalid. Note,
// the SigVerificationDecorator decorator will not get executed on ReCheck.
//
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
type SigVerificationDecorator struct {
	ak keeper.AccountKeeper
}

func NewSigVerificationDecorator(ak keeper.AccountKeeper) SigVerificationDecorator {
	return SigVerificationDecorator{
		ak: ak,
	}
}

func (svd SigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// no need to verify signatures on recheck tx
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}
	sigTx, ok := tx.(SigVerifiableTx)
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

// IncrementSequenceDecorator handles incrementing sequences of all signers.
// Use the IncrementSequenceDecorator decorator to prevent replay attacks. Note,
// there is no need to execute IncrementSequenceDecorator on CheckTx or RecheckTX
// since it is merely updating the nonce. As a result, this has the side effect
// that subsequent and sequential txs orginating from the same account cannot be
// handled correctly in a reliable way. To send sequential txs orginating from the
// same account, it is recommended to instead use multiple messages in a tx.
//
// CONTRACT: The tx must implement the SigVerifiableTx interface.
type IncrementSequenceDecorator struct {
	ak keeper.AccountKeeper
}

func NewIncrementSequenceDecorator(ak keeper.AccountKeeper) IncrementSequenceDecorator {
	return IncrementSequenceDecorator{
		ak: ak,
	}
}

func (isd IncrementSequenceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// no need to increment sequence on CheckTx or RecheckTx
	if ctx.IsCheckTx() && !simulate {
		return next(ctx, tx, simulate)
	}

	sigTx, ok := tx.(SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// increment sequence of all signers
	for _, addr := range sigTx.GetSigners() {
		acc := isd.ak.GetAccount(ctx, addr)
		if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
			panic(err)
		}
		isd.ak.SetAccount(ctx, acc)
	}

	return next(ctx, tx, simulate)
}

// ValidateSigCountDecorator takes in Params and returns errors if there are too many signatures in the tx for the given params
// otherwise it calls next AnteHandler
// Use this decorator to set parameterized limit on number of signatures in tx
// CONTRACT: Tx must implement SigVerifiableTx interface
type ValidateSigCountDecorator struct {
	ak keeper.AccountKeeper
}

func NewValidateSigCountDecorator(ak keeper.AccountKeeper) ValidateSigCountDecorator {
	return ValidateSigCountDecorator{
		ak: ak,
	}
}

func (vscd ValidateSigCountDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a sigTx")
	}

	params := vscd.ak.GetParams(ctx)
	pubKeys := sigTx.GetPubKeys()

	sigCount := 0
	for _, pk := range pubKeys {
		sigCount += types.CountSubKeys(pk)
		if uint64(sigCount) > params.TxSigLimit {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrTooManySignatures,
				"signatures: %d, limit: %d", sigCount, params.TxSigLimit)
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
		return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "ED25519 public keys are unsupported")

	case secp256k1.PubKeySecp256k1:
		meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: secp256k1")
		return nil

	case multisig.PubKeyMultisigThreshold:
		var multisignature multisig.Multisignature
		codec.Cdc.MustUnmarshalBinaryBare(sig, &multisignature)

		ConsumeMultisignatureVerificationGas(meter, multisignature, pubkey, params)
		return nil

	default:
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
	}
}

// ConsumeMultisignatureVerificationGas consumes gas from a GasMeter for verifying a multisig pubkey signature
func ConsumeMultisignatureVerificationGas(meter sdk.GasMeter,
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

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak keeper.AccountKeeper, addr sdk.AccAddress) (exported.Account, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}
	return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
}
