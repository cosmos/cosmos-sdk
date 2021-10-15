package middleware

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	// simulation signature values used to estimate gas consumption
	key                = make([]byte, secp256k1.PubKeySize)
	simSecp256k1Pubkey = &secp256k1.PubKey{Key: key}
	simSecp256k1Sig    [64]byte
)

// SignatureVerificationGasConsumer is the type of function that is used to both
// consume gas when verifying signatures and also to accept or reject different types of pubkeys
// This is where apps can define their own PubKey
type SignatureVerificationGasConsumer = func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error

var _ tx.Handler = setPubKeyTxHandler{}

type setPubKeyTxHandler struct {
	ak   AccountKeeper
	next tx.Handler
}

// SetPubKeyMiddleware sets PubKeys in context for any signer which does not already have pubkey set
// PubKeys must be set in context for all signers before any other sigverify middlewares run
// CONTRACT: Tx must implement SigVerifiableTx interface
func SetPubKeyMiddleware(ak AccountKeeper) tx.Middleware {
	return func(txh tx.Handler) tx.Handler {
		return setPubKeyTxHandler{
			ak:   ak,
			next: txh,
		}
	}
}

func (spkm setPubKeyTxHandler) setPubKey(ctx context.Context, tx sdk.Tx, simulate bool) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}

	pubkeys, err := sigTx.GetPubKeys()
	if err != nil {
		return err
	}
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
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey,
				"pubKey does not match signer address %s with signer index: %d", signers[i], i)
		}

		acc, err := GetSignerAcc(sdkCtx, spkm.ak, signers[i])
		if err != nil {
			return err
		}
		// account already has pubkey set,no need to reset
		if acc.GetPubKey() != nil {
			continue
		}
		err = acc.SetPubKey(pk)
		if err != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, err.Error())
		}
		spkm.ak.SetAccount(sdkCtx, acc)
	}

	// Also emit the following events, so that txs can be indexed by these
	// indices:
	// - signature (via `tx.signature='<sig_as_base64>'`),
	// - concat(address,"/",sequence) (via `tx.acc_seq='cosmos1abc...def/42'`).
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return err
	}

	var events sdk.Events
	for i, sig := range sigs {
		events = append(events, sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyAccountSequence, fmt.Sprintf("%s/%d", signers[i], sig.Sequence)),
		))

		sigBzs, err := signatureDataToBz(sig.Data)
		if err != nil {
			return err
		}
		for _, sigBz := range sigBzs {
			events = append(events, sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeySignature, base64.StdEncoding.EncodeToString(sigBz)),
			))
		}
	}

	sdkCtx.EventManager().EmitEvents(events)

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (spkm setPubKeyTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := spkm.setPubKey(ctx, tx, false); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return spkm.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (spkm setPubKeyTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := spkm.setPubKey(ctx, tx, false); err != nil {
		return abci.ResponseDeliverTx{}, err
	}
	return spkm.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (spkm setPubKeyTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := spkm.setPubKey(ctx, sdkTx, true); err != nil {
		return tx.ResponseSimulateTx{}, err
	}
	return spkm.next.SimulateTx(ctx, sdkTx, req)
}

var _ tx.Handler = validateSigCountTxHandler{}

type validateSigCountTxHandler struct {
	ak   AccountKeeper
	next tx.Handler
}

// ValidateSigCountMiddleware takes in Params and returns errors if there are too many signatures in the tx for the given params
// otherwise it calls next middleware
// Use this middleware to set parameterized limit on number of signatures in tx
// CONTRACT: Tx must implement SigVerifiableTx interface
func ValidateSigCountMiddleware(ak AccountKeeper) tx.Middleware {
	return func(txh tx.Handler) tx.Handler {
		return validateSigCountTxHandler{
			ak:   ak,
			next: txh,
		}
	}
}

func (vscd validateSigCountTxHandler) checkSigCount(ctx context.Context, tx sdk.Tx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a sigTx")
	}

	params := vscd.ak.GetParams(sdkCtx)
	pubKeys, err := sigTx.GetPubKeys()
	if err != nil {
		return err
	}

	sigCount := 0
	for _, pk := range pubKeys {
		sigCount += CountSubKeys(pk)
		if uint64(sigCount) > params.TxSigLimit {
			return sdkerrors.Wrapf(sdkerrors.ErrTooManySignatures,
				"signatures: %d, limit: %d", sigCount, params.TxSigLimit)
		}
	}
	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (vscd validateSigCountTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := vscd.checkSigCount(ctx, tx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return vscd.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (vscd validateSigCountTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := vscd.checkSigCount(ctx, sdkTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return vscd.next.SimulateTx(ctx, sdkTx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (vscd validateSigCountTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := vscd.checkSigCount(ctx, tx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return vscd.next.DeliverTx(ctx, tx, req)
}

// DefaultSigVerificationGasConsumer is the default implementation of SignatureVerificationGasConsumer. It consumes gas
// for signature verification based upon the public key type. The cost is fetched from the given params and is matched
// by the concrete type.
func DefaultSigVerificationGasConsumer(
	meter sdk.GasMeter, sig signing.SignatureV2, params types.Params,
) error {
	pubkey := sig.PubKey
	switch pubkey := pubkey.(type) {
	case *ed25519.PubKey:
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "ED25519 public keys are unsupported")

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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
	}
}

// ConsumeMultisignatureVerificationGas consumes gas from a GasMeter for verifying a multisig pubkey signature
func ConsumeMultisignatureVerificationGas(
	meter sdk.GasMeter, sig *signing.MultiSignatureData, pubkey multisig.PubKey,
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

var _ tx.Handler = sigGasConsumeTxHandler{}

type sigGasConsumeTxHandler struct {
	ak             AccountKeeper
	sigGasConsumer SignatureVerificationGasConsumer
	next           tx.Handler
}

// SigGasConsumeMiddleware consumes parameter-defined amount of gas for each signature according to the passed-in SignatureVerificationGasConsumer function
// before calling the next middleware
// CONTRACT: Pubkeys are set in context for all signers before this middleware runs
// CONTRACT: Tx must implement SigVerifiableTx interface
func SigGasConsumeMiddleware(ak AccountKeeper, sigGasConsumer SignatureVerificationGasConsumer) tx.Middleware {
	return func(h tx.Handler) tx.Handler {
		return sigGasConsumeTxHandler{
			ak:             ak,
			sigGasConsumer: sigGasConsumer,
			next:           h,
		}
	}
}

func (sgcm sigGasConsumeTxHandler) sigGasConsume(ctx context.Context, tx sdk.Tx, simulate bool) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	params := sgcm.ak.GetParams(sdkCtx)
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return err
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signerAddrs := sigTx.GetSigners()

	for i, sig := range sigs {
		signerAcc, err := GetSignerAcc(sdkCtx, sgcm.ak, signerAddrs[i])
		if err != nil {
			return err
		}

		pubKey := signerAcc.GetPubKey()

		// In simulate mode the transaction comes with no signatures, thus if the
		// account's pubkey is nil, both signature verification and gasKVStore.Set()
		// shall consume the largest amount, i.e. it takes more gas to verify
		// secp256k1 keys than ed25519 ones.
		if simulate && pubKey == nil {
			pubKey = simSecp256k1Pubkey
		}

		// make a SignatureV2 with PubKey filled in from above
		sig = signing.SignatureV2{
			PubKey:   pubKey,
			Data:     sig.Data,
			Sequence: sig.Sequence,
		}

		err = sgcm.sigGasConsumer(sdkCtx.GasMeter(), sig, params)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (sgcm sigGasConsumeTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := sgcm.sigGasConsume(ctx, tx, false); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return sgcm.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (sgcm sigGasConsumeTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := sgcm.sigGasConsume(ctx, tx, false); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return sgcm.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (sgcm sigGasConsumeTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := sgcm.sigGasConsume(ctx, sdkTx, true); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return sgcm.next.SimulateTx(ctx, sdkTx, req)
}

var _ tx.Handler = sigVerificationTxHandler{}

type sigVerificationTxHandler struct {
	ak              AccountKeeper
	signModeHandler authsigning.SignModeHandler
	next            tx.Handler
}

// SigVerificationMiddleware verifies all signatures for a tx and return an error if any are invalid. Note,
// the sigVerificationTxHandler middleware will not get executed on ReCheck.
//
// CONTRACT: Pubkeys are set in context for all signers before this middleware runs
// CONTRACT: Tx must implement SigVerifiableTx interface
func SigVerificationMiddleware(ak AccountKeeper, signModeHandler authsigning.SignModeHandler) tx.Middleware {
	return func(h tx.Handler) tx.Handler {
		return sigVerificationTxHandler{
			ak:              ak,
			signModeHandler: signModeHandler,
			next:            h,
		}
	}
}

// OnlyLegacyAminoSigners checks SignatureData to see if all
// signers are using SIGN_MODE_LEGACY_AMINO_JSON. If this is the case
// then the corresponding SignatureV2 struct will not have account sequence
// explicitly set, and we should skip the explicit verification of sig.Sequence
// in the SigVerificationMiddleware's middleware function.
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

func (svm sigVerificationTxHandler) sigVerify(ctx context.Context, tx sdk.Tx, isReCheckTx, simulate bool) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// no need to verify signatures on recheck tx
	if isReCheckTx {
		return nil
	}
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return err
	}

	signerAddrs := sigTx.GetSigners()

	// check that signer length and signature length are the same
	if len(sigs) != len(signerAddrs) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", len(signerAddrs), len(sigs))
	}

	for i, sig := range sigs {
		acc, err := GetSignerAcc(sdkCtx, svm.ak, signerAddrs[i])
		if err != nil {
			return err
		}

		// retrieve pubkey
		pubKey := acc.GetPubKey()
		if !simulate && pubKey == nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
		}

		// Check account sequence number.
		if sig.Sequence != acc.GetSequence() {
			return sdkerrors.Wrapf(
				sdkerrors.ErrWrongSequence,
				"account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence,
			)
		}

		// retrieve signer data
		genesis := sdkCtx.BlockHeight() == 0
		chainID := sdkCtx.ChainID()
		var accNum uint64
		if !genesis {
			accNum = acc.GetAccountNumber()
		}

		signerData := authsigning.SignerData{
			Address:       signerAddrs[i].String(),
			ChainID:       chainID,
			AccountNumber: accNum,
			Sequence:      acc.GetSequence(),
			SignerIndex:   i,
		}

		if !simulate {
			err := authsigning.VerifySignature(pubKey, signerData, sig.Data, svm.signModeHandler, tx)
			if err != nil {
				var errMsg string
				if OnlyLegacyAminoSigners(sig.Data) {
					// If all signers are using SIGN_MODE_LEGACY_AMINO, we rely on VerifySignature to check account sequence number,
					// and therefore communicate sequence number as a potential cause of error.
					errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d), sequence (%d) and chain-id (%s)", accNum, acc.GetSequence(), chainID)
				} else {
					errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d) and chain-id (%s)", accNum, chainID)
				}
				return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, errMsg)

			}
		}
	}

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (svd sigVerificationTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := svd.sigVerify(ctx, tx, req.Type == abci.CheckTxType_Recheck, false); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return svd.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (svd sigVerificationTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := svd.sigVerify(ctx, tx, false, false); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return svd.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (svd sigVerificationTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := svd.sigVerify(ctx, sdkTx, false, true); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return svd.next.SimulateTx(ctx, sdkTx, req)
}

var _ tx.Handler = incrementSequenceTxHandler{}

type incrementSequenceTxHandler struct {
	ak   AccountKeeper
	next tx.Handler
}

// IncrementSequenceMiddleware handles incrementing sequences of all signers.
// Use the incrementSequenceTxHandler middleware to prevent replay attacks. Note,
// there is no need to execute incrementSequenceTxHandler on RecheckTX since
// CheckTx would already bump the sequence number.
//
// NOTE: Since CheckTx and DeliverTx state are managed separately, subsequent and
// sequential txs orginating from the same account cannot be handled correctly in
// a reliable way unless sequence numbers are managed and tracked manually by a
// client. It is recommended to instead use multiple messages in a tx.
func IncrementSequenceMiddleware(ak AccountKeeper) tx.Middleware {
	return func(h tx.Handler) tx.Handler {
		return incrementSequenceTxHandler{
			ak:   ak,
			next: h,
		}
	}
}

func (isd incrementSequenceTxHandler) incrementSeq(ctx context.Context, tx sdk.Tx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// increment sequence of all signers
	for _, addr := range sigTx.GetSigners() {
		acc := isd.ak.GetAccount(sdkCtx, addr)
		if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
			panic(err)
		}

		isd.ak.SetAccount(sdkCtx, acc)
	}

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (isd incrementSequenceTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := isd.incrementSeq(ctx, tx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return isd.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (isd incrementSequenceTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := isd.incrementSeq(ctx, tx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return isd.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (isd incrementSequenceTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if err := isd.incrementSeq(ctx, sdkTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return isd.next.SimulateTx(ctx, sdkTx, req)
}

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak AccountKeeper, addr sdk.AccAddress) (types.AccountI, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}

	return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
}

// CountSubKeys counts the total number of keys for a multi-sig public key.
func CountSubKeys(pub cryptotypes.PubKey) int {
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

		multisig := cryptotypes.MultiSignature{
			Signatures: sigs,
		}
		aggregatedSig, err := multisig.Marshal()
		if err != nil {
			return nil, err
		}
		sigs = append(sigs, aggregatedSig)

		return sigs, nil
	default:
		return nil, sdkerrors.ErrInvalidType.Wrapf("unexpected signature data type %T", data)
	}
}
