package ante

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	secp256k1dcrd "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"google.golang.org/protobuf/types/known/anypb"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	authsigning "cosmossdk.io/x/auth/signing"
	"cosmossdk.io/x/auth/types"
	txsigning "cosmossdk.io/x/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var (
	// simulation signature values used to estimate gas consumption
	key                = make([]byte, secp256k1.PubKeySize)
	simSecp256k1Pubkey = &secp256k1.PubKey{Key: key}
	simSecp256k1Sig    [64]byte
)

func init() {
	// This decodes a valid hex string into a sepc256k1Pubkey for use in transaction simulation
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(key, bz)
	simSecp256k1Pubkey.Key = key
}

// SignatureVerificationGasConsumer is the type of function that is used to both
// consume gas when verifying signatures and also to accept or reject different types of pubkeys
// This is where apps can define their own PubKey
type SignatureVerificationGasConsumer = func(meter storetypes.GasMeter, sig signing.SignatureV2, params types.Params) error

// SetPubKeyDecorator sets PubKeys in context for any signer which does not already have pubkey set
// PubKeys must be set in context for all signers before any other sigverify decorators run
// CONTRACT: Tx must implement SigVerifiableTx interface
type SetPubKeyDecorator struct {
	ak AccountKeeper
}

func NewSetPubKeyDecorator(ak AccountKeeper) SetPubKeyDecorator {
	return SetPubKeyDecorator{
		ak: ak,
	}
}

func (spkd SetPubKeyDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}

	pubkeys, err := sigTx.GetPubKeys()
	if err != nil {
		return ctx, err
	}

	signers, err := sigTx.GetSigners()
	if err != nil {
		return sdk.Context{}, err
	}

	signerStrs := make([]string, len(signers))
	for i, pk := range pubkeys {
		var err error
		signerStrs[i], err = spkd.ak.AddressCodec().BytesToString(signers[i])
		if err != nil {
			return sdk.Context{}, err
		}

		// PublicKey was omitted from slice since it has already been set in context
		if pk == nil {
			if !simulate {
				continue
			}
			pk = simSecp256k1Pubkey
		}
		// Only make check if simulate=false
		if !simulate && !bytes.Equal(pk.Address(), signers[i]) && ctx.IsSigverifyTx() {
			return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidPubKey,
				"pubKey does not match signer address %s with signer index: %d", signerStrs[i], i)
		}
		if err := verifyIsOnCurve(pk); err != nil {
			return ctx, err
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
			return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, err.Error())
		}
		spkd.ak.SetAccount(ctx, acc)
	}

	// Also emit the following events, so that txs can be indexed by these
	// indices:
	// - signature (via `tx.signature='<sig_as_base64>'`),
	// - concat(address,"/",sequence) (via `tx.acc_seq='cosmos1abc...def/42'`).
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return ctx, err
	}

	var events sdk.Events
	for i, sig := range sigs {
		events = append(events, sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyAccountSequence, fmt.Sprintf("%s/%d", signerStrs[i], sig.Sequence)),
		))

		sigBzs, err := signatureDataToBz(sig.Data)
		if err != nil {
			return ctx, err
		}
		for _, sigBz := range sigBzs {
			events = append(events, sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeySignature, base64.StdEncoding.EncodeToString(sigBz)),
			))
		}
	}

	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

// SigVerificationDecorator verifies all signatures for a tx and returns an
// error if any are invalid. Note, the SigVerificationDecorator will not check
// signatures on ReCheckTx. It will also increase the sequence number, and consume
// gas for signature verification.
//
// In cases where unordered or parallel transactions are desired, it is recommended
// to to set unordered=true with a reasonable timeout_height value, in which case
// this nonce verification and increment will be skipped.
//
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
type SigVerificationDecorator struct {
	ak              AccountKeeper
	signModeHandler *txsigning.HandlerMap
	sigGasConsumer  SignatureVerificationGasConsumer
}

func NewSigVerificationDecorator(ak AccountKeeper, signModeHandler *txsigning.HandlerMap, sigGasConsumer SignatureVerificationGasConsumer) SigVerificationDecorator {
	return SigVerificationDecorator{
		ak:              ak,
		signModeHandler: signModeHandler,
		sigGasConsumer:  sigGasConsumer,
	}
}

// OnlyLegacyAminoSigners checks SignatureData to see if all
// signers are using SIGN_MODE_LEGACY_AMINO_JSON. If this is the case
// then the corresponding SignatureV2 struct will not have account sequence
// explicitly set, and we should skip the explicit verification of sig.Sequence
// in the SigVerificationDecorator's AnteHandler function.
func OnlyLegacyAminoSigners(sigData signing.SignatureData) bool {
	switch v := sigData.(type) {
	case *signing.SingleSignatureData:
		return v.SignMode == signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	case *signing.MultiSignatureData:
		for _, s := range v.Signatures {
			if !OnlyLegacyAminoSigners(s) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func verifyIsOnCurve(pubKey cryptotypes.PubKey) (err error) {
	switch typedPubKey := pubKey.(type) {
	case *secp256k1.PubKey:
		pubKeyObject, err := secp256k1dcrd.ParsePubKey(typedPubKey.Bytes())
		if err != nil {
			if errors.Is(err, secp256k1dcrd.ErrPubKeyNotOnCurve) {
				return errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "secp256k1 key is not on curve")
			}
			return err
		}
		if !pubKeyObject.IsOnCurve() {
			return errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "secp256k1 key is not on curve")
		}

	case *secp256r1.PubKey:
		pubKeyObject := typedPubKey.Key.PublicKey
		if !pubKeyObject.IsOnCurve(pubKeyObject.X, pubKeyObject.Y) {
			return errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "secp256r1 key is not on curve")
		}

	case multisig.PubKey:
		pubKeysObjects := typedPubKey.GetPubKeys()
		ok := true
		for _, pubKeyObject := range pubKeysObjects {
			if err := verifyIsOnCurve(pubKeyObject); err != nil {
				ok = false
				break
			}
		}
		if !ok {
			return errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "some keys are not on curve")
		}

	default:
		return errorsmod.Wrapf(sdkerrors.ErrInvalidPubKey, "unsupported key type: %T", typedPubKey)
	}

	return nil
}

func (svd SigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signatures, err := sigTx.GetSignaturesV2()
	if err != nil {
		return ctx, err
	}

	signers, err := sigTx.GetSigners()
	if err != nil {
		return ctx, err
	}

	// check that signer length and signature length are the same
	if len(signatures) != len(signers) {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", len(signers), len(signatures))
	}

	for i := range signers {
		err = svd.authenticate(ctx, tx, simulate, signers[i], signatures[i])
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// authenticate the authentication of the TX for a specific tx signer.
func (svd SigVerificationDecorator) authenticate(ctx sdk.Context, tx sdk.Tx, simulate bool, signer []byte, sig signing.SignatureV2) error {
	acc, err := GetSignerAcc(ctx, svd.ak, signer)
	if err != nil {
		return err
	}

	err = svd.consumeSignatureGas(ctx, simulate, acc.GetPubKey(), sig)
	if err != nil {
		return err
	}

	err = svd.verifySig(ctx, simulate, tx, acc, sig)
	if err != nil {
		return err
	}

	// Bypass incrementing sequence for transactions with unordered set to true.
	// The actual parameters of the un-ordered tx will be checked in a separate
	// decorator.
	unorderedTx, ok := tx.(sdk.TxWithUnordered)
	if ok && unorderedTx.GetUnordered() {
		return nil
	}

	return svd.increaseSequence(ctx, acc)
}

// consumeSignatureGas will consume gas according to the pub-key being verified.
func (svd SigVerificationDecorator) consumeSignatureGas(
	ctx sdk.Context,
	simulate bool,
	pubKey cryptotypes.PubKey,
	signature signing.SignatureV2,
) error {
	if simulate && pubKey == nil {
		pubKey = simSecp256k1Pubkey
	}

	// make a SignatureV2 with PubKey filled in from above
	signature = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     signature.Data,
		Sequence: signature.Sequence,
	}

	err := svd.sigGasConsumer(ctx.GasMeter(), signature, svd.ak.GetParams(ctx))
	if err != nil {
		return err
	}
	return nil
}

// verifySig will verify the signature of the provided signer account.
// it will assess:
// - the pub key is on the curve.
// - verify sig
func (svd SigVerificationDecorator) verifySig(ctx sdk.Context, simulate bool, tx sdk.Tx, acc sdk.AccountI, sig signing.SignatureV2) error {
	// retrieve pubkey
	pubKey := acc.GetPubKey()
	if !simulate && pubKey == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
	}

	if err := verifyIsOnCurve(pubKey); err != nil {
		return err
	}

	if sig.Sequence != acc.GetSequence() {
		return errorsmod.Wrapf(
			sdkerrors.ErrWrongSequence,
			"account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence,
		)
	}

	// we're in simulation mode, or in ReCheckTx, or context is not
	// on sig verify tx, then we do not need to verify the signatures
	// in the tx.
	if simulate || ctx.IsReCheckTx() || !ctx.IsSigverifyTx() {
		return nil
	}

	// retrieve signer data
	genesis := ctx.BlockHeight() == 0
	chainID := ctx.ChainID()
	var accNum uint64
	if !genesis {
		accNum = acc.GetAccountNumber()
	}

	anyPk, _ := codectypes.NewAnyWithValue(pubKey)

	signerData := txsigning.SignerData{
		Address:       acc.GetAddress().String(),
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      acc.GetSequence(),
		PubKey: &anypb.Any{
			TypeUrl: anyPk.TypeUrl,
			Value:   anyPk.Value,
		},
	}
	adaptableTx, ok := tx.(authsigning.V2AdaptableTx)
	if !ok {
		return fmt.Errorf("expected tx to implement V2AdaptableTx, got %T", tx)
	}
	txData := adaptableTx.GetSigningTxData()
	err := authsigning.VerifySignature(ctx, pubKey, signerData, sig.Data, svd.signModeHandler, txData)
	if err != nil {
		var errMsg string
		if OnlyLegacyAminoSigners(sig.Data) {
			// If all signers are using SIGN_MODE_LEGACY_AMINO, we rely on VerifySignature to check account sequence number,
			// and therefore communicate sequence number as a potential cause of error.
			errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d), sequence (%d) and chain-id (%s)", accNum, acc.GetSequence(), chainID)
		} else {
			errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d) and chain-id (%s): (%s)", accNum, chainID, err.Error())
		}
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, errMsg)
	}

	return nil
}

// increaseSequence will increase the sequence number of the account.
func (svd SigVerificationDecorator) increaseSequence(ctx sdk.Context, acc sdk.AccountI) error {
	if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
		return err
	}

	svd.ak.SetAccount(ctx, acc)
	return nil
}

// ValidateSigCountDecorator takes in Params and returns errors if there are too many signatures in the tx for the given params
// otherwise it calls next AnteHandler
// Use this decorator to set parameterized limit on number of signatures in tx
// CONTRACT: Tx must implement SigVerifiableTx interface
type ValidateSigCountDecorator struct {
	ak AccountKeeper
}

func NewValidateSigCountDecorator(ak AccountKeeper) ValidateSigCountDecorator {
	return ValidateSigCountDecorator{
		ak: ak,
	}
}

func (vscd ValidateSigCountDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a sigTx")
	}

	params := vscd.ak.GetParams(ctx)
	pubKeys, err := sigTx.GetPubKeys()
	if err != nil {
		return ctx, err
	}

	sigCount := 0
	for _, pk := range pubKeys {
		sigCount += CountSubKeys(pk)
		if uint64(sigCount) > params.TxSigLimit {
			return ctx, errorsmod.Wrapf(sdkerrors.ErrTooManySignatures, "signatures: %d, limit: %d", sigCount, params.TxSigLimit)
		}
	}

	return next(ctx, tx, simulate)
}

// DefaultSigVerificationGasConsumer is the default implementation of SignatureVerificationGasConsumer. It consumes gas
// for signature verification based upon the public key type. The cost is fetched from the given params and is matched
// by the concrete type.
func DefaultSigVerificationGasConsumer(meter storetypes.GasMeter, sig signing.SignatureV2, params types.Params) error {
	pubkey := sig.PubKey

	switch pubkey := pubkey.(type) {
	case *ed25519.PubKey:
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "ED25519 public keys are unsupported")

	case *secp256k1.PubKey:
		meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: secp256k1")
		return nil

	case *secp256r1.PubKey:
		meter.ConsumeGas(params.SigVerifyCostSecp256r1(), "ante verify: secp256r1")
		return nil

	case multisig.PubKey:
		multisignature, ok := sig.Data.(*signing.MultiSignatureData)
		if !ok {
			return fmt.Errorf("expected %T, got, %T", &signing.MultiSignatureData{}, sig.Data)
		}

		err := ConsumeMultisignatureVerificationGas(meter, multisignature, pubkey, params, sig.Sequence)
		if err != nil {
			return err
		}

		return nil

	default:
		return errorsmod.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
	}
}

// ConsumeMultisignatureVerificationGas consumes gas from a GasMeter for verifying a multisig pubkey signature
func ConsumeMultisignatureVerificationGas(
	meter storetypes.GasMeter, sig *signing.MultiSignatureData, pubkey multisig.PubKey,
	params types.Params, accSeq uint64,
) error {
	size := sig.BitArray.Count()
	sigIndex := 0

	for i := 0; i < size; i++ {
		if !sig.BitArray.GetIndex(i) {
			continue
		}
		sigV2 := signing.SignatureV2{
			PubKey:   pubkey.GetPubKeys()[i],
			Data:     sig.Signatures[sigIndex],
			Sequence: accSeq,
		}

		err := DefaultSigVerificationGasConsumer(meter, sigV2, params)
		if err != nil {
			return err
		}

		sigIndex++
	}

	return nil
}

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak AccountKeeper, addr sdk.AccAddress) (sdk.AccountI, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}

	return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
}

// CountSubKeys counts the total number of keys for a multi-sig public key.
// A non-multisig, i.e. a regular signature, it naturally a count of 1. If it is a multisig,
// then it recursively calls it on its pubkeys.
func CountSubKeys(pub cryptotypes.PubKey) int {
	if pub == nil {
		return 0
	}

	v, ok := pub.(*kmultisig.LegacyAminoPubKey)
	if !ok {
		return 1
	}

	numKeys := 0
	for _, subkey := range v.GetPubKeys() {
		numKeys += CountSubKeys(subkey)
	}

	return numKeys
}

// signatureDataToBz converts a SignatureData into raw bytes signature.
// For SingleSignatureData, it returns the signature raw bytes.
// For MultiSignatureData, it returns an array of all individual signatures,
// as well as the aggregated signature.
func signatureDataToBz(data signing.SignatureData) ([][]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("got empty SignatureData")
	}

	switch data := data.(type) {
	case *signing.SingleSignatureData:
		return [][]byte{data.Signature}, nil

	case *signing.MultiSignatureData:
		sigs := [][]byte{}
		var err error

		for _, d := range data.Signatures {
			nestedSigs, err := signatureDataToBz(d)
			if err != nil {
				return nil, err
			}

			sigs = append(sigs, nestedSigs...)
		}

		multiSignature := cryptotypes.MultiSignature{
			Signatures: sigs,
		}

		aggregatedSig, err := multiSignature.Marshal()
		if err != nil {
			return nil, err
		}

		sigs = append(sigs, aggregatedSig)
		return sigs, nil

	default:
		return nil, sdkerrors.ErrInvalidType.Wrapf("unexpected signature data type %T", data)
	}
}
