package secp256k1

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/codec"
	secp "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1/internal/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ cryptotypes.PrivKey  = &EthPrivKey{}
	_ codec.AminoMarshaler = &EthPrivKey{}
)

const (
	EthPrivKeySize     = 32
	EthPubKeySize      = EthPrivKeySize + 1
	ethkeyType         = "ethsecp256k1"
	EthPrivKeyName     = "eth/PrivKeySecp256k1"
	EthPubKeyName      = "eth/PubKeySecp256k1"
	RecoveryIDOffset   = 64
	EthSignatureLength = RecoveryIDOffset + 1
	DigestLength       = 32
	wordBits           = 32 << (uint64(^big.Word(0)) >> 63)
	wordBytes          = wordBits / 8
)

var (
	secp256k1N, _ = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
)

// Amino encoding names
const ()

// ----------------------------------------------------------------------------
// secp256k1 Private Key

// GenerateKey generates a new random private key. It returns an error upon
// failure.
func GenerateKey() (*EthPrivKey, error) {
	priv, err := ecdsa.GenerateKey(secp.S256(), rand.Reader)
	// priv, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	return &EthPrivKey{
		Key: FromECDSA(priv),
	}, nil
}

// Bytes returns the byte representation of the ECDSA Private Key.
func (privKey EthPrivKey) Bytes() []byte {
	bz := make([]byte, len(privKey.Key))
	copy(bz, privKey.Key)

	return bz
}

// EthPubKey returns the ECDSA private key's public key. If the privkey is not valid
// it returns a nil value.
func (privKey EthPrivKey) PubKey() cryptotypes.PubKey {
	ecdsaPrivKey, err := privKey.ToECDSA()
	if err != nil {
		return nil
	}

	return &EthPubKey{
		Key: secp.CompressPubkey(ecdsaPrivKey.PublicKey.X, ecdsaPrivKey.PublicKey.Y),
	}
}

// Equals returns true if two ECDSA private keys are equal and false otherwise.
func (privKey EthPrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && subtle.ConstantTimeCompare(privKey.Bytes(), other.Bytes()) == 1
}

// Type returns eth_secp256k1
func (privKey EthPrivKey) Type() string {
	return ethkeyType
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey EthPrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *EthPrivKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PrivKeySize {
		return fmt.Errorf("invalid privkey size, expected %d got %d", PrivKeySize, len(bz))
	}
	privKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (privKey EthPrivKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return privKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (privKey *EthPrivKey) UnmarshalAminoJSON(bz []byte) error {
	return privKey.UnmarshalAmino(bz)
}

// Sign creates a recoverable ECDSA signature on the secp256k1 curve over the
// provided hash of the message. The produced signature is 65 bytes
// where the last byte contains the recovery ID.
func (privKey EthPrivKey) Sign(digestBz []byte) ([]byte, error) {
	if len(digestBz) != DigestLength {
		digestBz = Keccak256(digestBz)
	}

	key, err := privKey.ToECDSA()
	if err != nil {
		return nil, err
	}

	if len(digestBz) != DigestLength {
		return nil, fmt.Errorf("hash is required to be exactly %d bytes (%d)", DigestLength, len(digestBz))
	}
	seckey := FromECDSA(key)
	defer func() {
		for i := range seckey {
			seckey[i] = 0
		}
	}()
	return secp.Sign(digestBz, seckey)
}

// ToECDSA returns the ECDSA private key as a reference to ecdsa.PrivateKey type.
func (privKey EthPrivKey) ToECDSA() (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = secp.S256()
	if 8*len(privKey.Key) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(privKey.Key)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(privKey.Key)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

// ----------------------------------------------------------------------------
// secp256k1 Public Key

var (
	_ cryptotypes.PubKey   = &EthPubKey{}
	_ codec.AminoMarshaler = &EthPubKey{}
)

// Address returns the address of the ECDSA public key.
// The function will return an empty address if the public key is invalid.
func (pubKey EthPubKey) Address() cmtcrypto.Address {
	pubk, err := DecompressPubkey(pubKey.Key)
	if err != nil || pubk == nil {
		return nil
	}
	pubBytes := FromECDSAPub(pubk)
	return cmtcrypto.Address(Keccak256(pubBytes[1:])[12:])
}

// Bytes returns the raw bytes of the ECDSA public key.
func (pubKey EthPubKey) Bytes() []byte {
	bz := make([]byte, len(pubKey.Key))
	copy(bz, pubKey.Key)

	return bz
}

// String implements the fmt.Stringer interface.
func (pubKey EthPubKey) String() string {
	return fmt.Sprintf("EthPubKeySecp256k1{%X}", pubKey.Key)
}

// Type returns eth_secp256k1
func (pubKey EthPubKey) Type() string {
	return ethkeyType
}

// Equals returns true if the pubkey type is the same and their bytes are deeply equal.
func (pubKey EthPubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// MarshalAmino overrides Amino binary marshaling.
func (pubKey EthPubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (pubKey *EthPubKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PubKeySize {
		return errorsmod.Wrapf(errortypes.ErrInvalidPubKey, "invalid pubkey size, expected %d, got %d", PubKeySize, len(bz))
	}
	pubKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey EthPubKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return pubKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey *EthPubKey) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}

// VerifySignature verifies that the ECDSA public key created a given signature over
// the provided message. It will calculate the Keccak256 hash of the message
// prior to verification and approve verification if the signature can be verified
// from the original message.
//
// CONTRACT: The signature should be in [R || S] format.
func (pubKey EthPubKey) VerifySignature(msg, sig []byte) bool {
	return pubKey.verifySignatureECDSA(msg, sig)
}

// Perform standard ECDSA signature verification for the given raw bytes and signature.
func (pubKey EthPubKey) verifySignatureECDSA(msg, sig []byte) bool {
	if len(sig) == EthSignatureLength {
		// remove recovery ID (V) if contained in the signature
		sig = sig[:len(sig)-1]
	}

	// the signature needs to be in [R || S] format when provided to VerifySignature
	return secp.VerifySignature(pubKey.Key, Keccak256(msg), sig)
}

// ----------------------------------------------------------------------------
// Helper Functions

func DecompressPubkey(pubkey []byte) (*ecdsa.PublicKey, error) {
	x, y := secp.DecompressPubkey(pubkey)
	if x == nil {
		return nil, fmt.Errorf("invalid public key")
	}
	return &ecdsa.PublicKey{X: x, Y: y, Curve: secp.S256()}, nil
}

func paddedBigBytes(bigint *big.Int, n int) []byte {
	if bigint.BitLen()/8 >= n {
		return bigint.Bytes()
	}
	ret := make([]byte, n)
	i := len(ret)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			ret[i] = byte(d)
			d >>= 8
		}
	}
	return ret
}

func FromECDSA(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return paddedBigBytes(priv.D, priv.Params().BitSize/8)
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp.S256(), pub.X, pub.Y)
}
