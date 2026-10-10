package types

import (
	"encoding/base64"
	"encoding/json"

	mldsa "github.com/cometbft/cometbft/crypto/mldsa65"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	multisigtypes "github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// ValidateRekeyPubKey checks that next can safely replace current as an
// account's pubkey. It accepts the key types the default signature
// verification gas consumer accepts (secp256k1, secp256r1, ed25519, mldsa65)
// and LegacyAminoPubKey multisigs built from them, except that secp256r1 is
// accepted only as a top-level key: kmultisig.AminoCdc does not register it,
// so a multisig holding one panics in Address(). A multisig must have a
// threshold in [1, number of subkeys], at most MaxSignatureTreeBreadth direct
// subkeys and at most params.TxSigLimit subkeys in total, and multisigs may
// nest at most MaxSignatureTreeDepth levels deep (a multisig may contain
// multisigs, but those may not). These limits match what the ante handler
// accepts, so the account cannot lock itself out.
// As with LegacyAminoPubKey elsewhere, duplicate subkeys are allowed and each
// slot counts toward the threshold. next must differ from current.
//
// Leaf keys are checked for length only, not for being valid curve points or
// encodings, so passing this check does not make next usable. Callers must
// also verify a proof with VerifyChangePubKeyProof. Callers must run this
// check first: ChangePubKeyProofSignBytes and VerifyChangePubKeyProof call
// Address() and GetPubKeys(), which panic on malformed or un-unpacked keys.
func ValidateRekeyPubKey(current, next cryptotypes.PubKey, params Params) error {
	if next == nil {
		return errorsmod.Wrap(ErrInvalidNewPubKey, "new pubkey is nil")
	}
	if err := validateRekeyKeyType(next, 0); err != nil {
		return err
	}
	if n := countRekeySubKeys(next); uint64(n) > params.TxSigLimit {
		return errorsmod.Wrapf(ErrInvalidNewPubKey, "new pubkey has %d subkeys, more than the tx sig limit %d", n, params.TxSigLimit)
	}
	if current != nil && next.Equals(current) {
		return errorsmod.Wrap(ErrInvalidNewPubKey, "new pubkey equals the current pubkey")
	}
	return nil
}

// validateRekeyKeyType recursively checks the type and well-formedness of pk.
// Key lengths are checked here because Address() panics on malformed keys.
// depth is the number of multisigs pk is nested in. A multisig subkey
// (depth > 0) is amino-encoded by LegacyAminoPubKey.Address() with
// kmultisig.AminoCdc, so it must be a type that codec registers.
func validateRekeyKeyType(pk cryptotypes.PubKey, depth int) error {
	switch pk := pk.(type) {
	case *secp256k1.PubKey:
		if pk == nil || len(pk.Key) != secp256k1.PubKeySize {
			return errorsmod.Wrap(ErrInvalidNewPubKey, "malformed secp256k1 pubkey")
		}
	case *ed25519.PubKey:
		if pk == nil || len(pk.Key) != ed25519.PubKeySize {
			return errorsmod.Wrap(ErrInvalidNewPubKey, "malformed ed25519 pubkey")
		}
	case *mldsa65.PubKey:
		if pk == nil || len(pk.Key) != mldsa.PubKeySize {
			return errorsmod.Wrap(ErrInvalidNewPubKey, "malformed mldsa65 pubkey")
		}
	case *secp256r1.PubKey:
		if pk == nil || pk.Key == nil {
			return errorsmod.Wrap(ErrInvalidNewPubKey, "malformed secp256r1 pubkey")
		}
		if depth > 0 {
			return errorsmod.Wrap(ErrInvalidNewPubKey, "secp256r1 is not supported as a multisig subkey")
		}
	case *kmultisig.LegacyAminoPubKey:
		if pk == nil {
			return errorsmod.Wrap(ErrInvalidNewPubKey, "nil multisig pubkey")
		}
		if depth >= MaxSignatureTreeDepth {
			return errorsmod.Wrapf(ErrInvalidNewPubKey, "multisig pubkey nested deeper than %d levels", MaxSignatureTreeDepth)
		}
		if len(pk.PubKeys) == 0 {
			return errorsmod.Wrap(ErrInvalidNewPubKey, "multisig pubkey has no subkeys")
		}
		if len(pk.PubKeys) > MaxSignatureTreeBreadth {
			return errorsmod.Wrapf(ErrInvalidNewPubKey, "multisig pubkey has %d subkeys, more than %d", len(pk.PubKeys), MaxSignatureTreeBreadth)
		}
		if pk.Threshold == 0 || uint64(pk.Threshold) > uint64(len(pk.PubKeys)) {
			return errorsmod.Wrapf(ErrInvalidNewPubKey, "multisig threshold %d is not in [1, %d]", pk.Threshold, len(pk.PubKeys))
		}
		// Read the cached values directly: GetPubKeys panics on an Any that
		// was never unpacked.
		for _, anyPk := range pk.PubKeys {
			if anyPk == nil {
				return errorsmod.Wrap(ErrInvalidNewPubKey, "multisig pubkey has a nil subkey")
			}
			sub, ok := anyPk.GetCachedValue().(cryptotypes.PubKey)
			if !ok || sub == nil {
				return errorsmod.Wrap(ErrInvalidNewPubKey, "multisig subkey is not an unpacked pubkey")
			}
			if err := validateRekeyKeyType(sub, depth+1); err != nil {
				return err
			}
		}
	default:
		return errorsmod.Wrapf(ErrInvalidNewPubKey, "unsupported pubkey type %T", pk)
	}
	return nil
}

// countRekeySubKeys counts the leaf keys of pk. It mirrors ante.CountSubKeys,
// which cannot be imported here without an import cycle.
func countRekeySubKeys(pk cryptotypes.PubKey) int {
	msig, ok := pk.(*kmultisig.LegacyAminoPubKey)
	if !ok {
		return 1
	}
	n := 0
	for _, sub := range msig.GetPubKeys() {
		n += countRekeySubKeys(sub)
	}
	return n
}

// adr036SignDoc is the ADR-036 off-chain sign doc: an amino-JSON StdSignDoc
// with empty chain id, zero account number, sequence and fee, and a single
// sign/MsgSignData message. Fields of every adr036* struct are declared in
// lexicographic JSON-key order so encoding/json emits canonical (sorted) amino
// JSON directly; TestChangePubKeyProofSignBytes_Golden pins the output.
type adr036SignDoc struct {
	AccountNumber string          `json:"account_number"`
	ChainID       string          `json:"chain_id"`
	Fee           adr036Fee       `json:"fee"`
	Memo          string          `json:"memo"`
	Msgs          []adr036Message `json:"msgs"`
	Sequence      string          `json:"sequence"`
}

type adr036Fee struct {
	Amount []struct{} `json:"amount"`
	Gas    string     `json:"gas"`
}

type adr036Message struct {
	Type  string            `json:"type"`
	Value adr036MsgSignData `json:"value"`
}

type adr036MsgSignData struct {
	Data   string `json:"data"`
	Signer string `json:"signer"`
}

// ChangePubKeyProofSignBytes returns the bytes the new key signs to prove
// possession: an ADR-036 amino-JSON sign doc whose signer is the new key's
// natural address and whose data is the proto encoding of doc. Binding the
// chain id, account number and address in doc stops a proof from being
// replayed for another account or chain.
func ChangePubKeyProofSignBytes(addrCodec address.Codec, doc ChangePubKeyProofDoc, newPk cryptotypes.PubKey) ([]byte, error) {
	if newPk == nil {
		return nil, errorsmod.Wrap(ErrInvalidNewPubKey, "new pubkey is nil")
	}
	signer, err := addrCodec.BytesToString(newPk.Address())
	if err != nil {
		return nil, err
	}
	data, err := doc.Marshal()
	if err != nil {
		return nil, err
	}
	bz, err := json.Marshal(adr036SignDoc{
		AccountNumber: "0",
		ChainID:       "",
		Fee:           adr036Fee{Amount: []struct{}{}, Gas: "0"},
		Memo:          "",
		Msgs: []adr036Message{{
			Type: "sign/MsgSignData",
			Value: adr036MsgSignData{
				Data:   base64.StdEncoding.EncodeToString(data),
				Signer: signer,
			},
		}},
		Sequence: "0",
	})
	if err != nil {
		return nil, err
	}
	return bz, nil
}

// Limits on the shape of a signature tree that the ante handler can flatten
// (see flattenSignatures in x/auth/ante). The top-level SignatureData is at
// depth 0 and data deeper than MaxSignatureTreeDepth is rejected, so a
// signature can have at most MaxSignatureTreeDepth levels of
// MultiSignatureData: a multisig may contain multisigs, but those may not.
// A MultiSignatureData may hold at most MaxSignatureTreeBreadth signatures.
// These values are consensus-critical; changing them changes which txs are
// valid.
const (
	MaxSignatureTreeDepth   = 2
	MaxSignatureTreeBreadth = 32
)

// MaxChangePubKeyProofDepth is the most MultiSignatureData levels a proof may
// nest. It matches MaxSignatureTreeDepth, since ValidateRekeyPubKey rejects
// any new pubkey the ante handler could not sign for with a deeper tree.
const MaxChangePubKeyProofDepth = MaxSignatureTreeDepth

// DecodeChangePubKeyProof decodes a MsgChangePubKey proof, the proto encoding
// of a SignatureDescriptor.Data.
func DecodeChangePubKeyProof(bz []byte) (signing.SignatureData, error) {
	var data signing.SignatureDescriptor_Data
	if err := data.Unmarshal(bz); err != nil {
		return nil, errorsmod.Wrap(ErrInvalidPubKeyProof, err.Error())
	}
	// SignatureDataFromProto panics on an unset oneof, so check the tree first.
	if err := validateSignatureDescriptorData(&data, 0); err != nil {
		return nil, err
	}
	return signing.SignatureDataFromProto(&data), nil
}

func validateSignatureDescriptorData(data *signing.SignatureDescriptor_Data, depth int) error {
	if data == nil {
		return errorsmod.Wrap(ErrInvalidPubKeyProof, "empty signature data")
	}
	switch sum := data.Sum.(type) {
	case *signing.SignatureDescriptor_Data_Single_:
		if sum.Single == nil {
			return errorsmod.Wrap(ErrInvalidPubKeyProof, "empty single signature")
		}
	case *signing.SignatureDescriptor_Data_Multi_:
		if depth >= MaxChangePubKeyProofDepth {
			return errorsmod.Wrapf(ErrInvalidPubKeyProof, "multisignature nested deeper than %d levels", MaxChangePubKeyProofDepth)
		}
		if sum.Multi == nil || sum.Multi.Bitarray == nil {
			return errorsmod.Wrap(ErrInvalidPubKeyProof, "multisignature without bit array")
		}
		for _, sub := range sum.Multi.Signatures {
			if err := validateSignatureDescriptorData(sub, depth+1); err != nil {
				return err
			}
		}
	default:
		return errorsmod.Wrap(ErrInvalidPubKeyProof, "empty signature data")
	}
	return nil
}

// VerifyChangePubKeyProof checks that proof is a valid signature by newPk over
// signBytes. A single key needs a SingleSignatureData and a multisig key needs
// a MultiSignatureData meeting its threshold. Every signature must use
// SIGN_MODE_LEGACY_AMINO_JSON, since signBytes is an ADR-036 amino-JSON doc.
func VerifyChangePubKeyProof(newPk cryptotypes.PubKey, signBytes []byte, proof signing.SignatureData) error {
	if newPk == nil || proof == nil {
		return errorsmod.Wrap(ErrInvalidPubKeyProof, "missing pubkey or proof")
	}
	getSignBytes := func(mode signing.SignMode) ([]byte, error) {
		if mode != signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
			return nil, errorsmod.Wrapf(ErrInvalidPubKeyProof, "sign mode %s, expected %s", mode, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
		}
		return signBytes, nil
	}

	switch sig := proof.(type) {
	case *signing.SingleSignatureData:
		if _, isMulti := newPk.(multisigtypes.PubKey); isMulti {
			return errorsmod.Wrap(ErrInvalidPubKeyProof, "multisig pubkey needs a multisignature proof")
		}
		msg, err := getSignBytes(sig.SignMode)
		if err != nil {
			return err
		}
		if !newPk.VerifySignature(msg, sig.Signature) {
			return errorsmod.Wrap(ErrInvalidPubKeyProof, "signature verification failed")
		}
		return nil
	case *signing.MultiSignatureData:
		msigPk, ok := newPk.(multisigtypes.PubKey)
		if !ok {
			return errorsmod.Wrapf(ErrInvalidPubKeyProof, "multisignature proof for non-multisig pubkey %T", newPk)
		}
		if err := msigPk.VerifyMultisignature(getSignBytes, sig); err != nil {
			return errorsmod.Wrap(ErrInvalidPubKeyProof, err.Error())
		}
		return nil
	default:
		return errorsmod.Wrapf(ErrInvalidPubKeyProof, "unsupported proof type %T", proof)
	}
}

var _ codectypes.UnpackInterfacesMessage = (*PubKeyHistoryEntry)(nil)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (e *PubKeyHistoryEntry) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if e == nil || e.PubKey == nil {
		return nil
	}
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(e.PubKey, &pk)
}

var _ codectypes.UnpackInterfacesMessage = (*GenesisPubKeyHistory)(nil)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (h *GenesisPubKeyHistory) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if h == nil {
		return nil
	}
	for i := range h.Entries {
		if err := h.Entries[i].UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

var _ codectypes.UnpackInterfacesMessage = (*ChangePubKeyProofDoc)(nil)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (d *ChangePubKeyProofDoc) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if d == nil || d.NewPubKey == nil {
		return nil
	}
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(d.NewPubKey, &pk)
}

// Events emitted by MsgChangePubKey.
const (
	EventTypeChangePubKey = "change_pubkey"

	AttributeKeyAddress          = "address"
	AttributeKeyOldPubKeyAddress = "old_pubkey_address"
	AttributeKeyNewPubKeyAddress = "new_pubkey_address"
)

var _ codectypes.UnpackInterfacesMessage = (*MsgChangePubKey)(nil)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces. It
// unpacks NewPubKey, including the subkeys of a multisig key.
func (m *MsgChangePubKey) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if m == nil || m.NewPubKey == nil {
		return nil
	}
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(m.NewPubKey, &pk)
}
