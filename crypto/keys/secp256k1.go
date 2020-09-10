package keys

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

// asserting interface implementation
var (
	_ crypto.PubKey        = &Secp256K1PubKey{}
	_ crypto.PrivKey       = &Secp256K1PrivKey{}
	_ codec.AminoMarshaler = &Secp256K1PrivKey{}
)

func (m *Secp256K1PubKey) Address() crypto.Address {
	return m.Key.Address()
}

func (m *Secp256K1PubKey) Bytes() []byte {
	return m.Key.Bytes()
}

func (m *Secp256K1PubKey) VerifySignature(msg []byte, sig []byte) bool {
	return m.Key.VerifySignature(msg, sig)
}

func (m *Secp256K1PubKey) Equals(key tmcrypto.PubKey) bool {
	return m.Key.Equals(key)
}

func (m *Secp256K1PubKey) Type() string {
	return m.Key.Type()
}

// MarshalAminoJSON overrides Amino binary marshalling.
func (m *Secp256K1PubKey) MarshalAmino() ([]byte, error) {
	return m.Key.Bytes(), nil
}

// UnmarshalAminoJSON overrides Amino binary marshalling.
func (m *Secp256K1PubKey) UnmarshalAmino(bz []byte) error {
	*m = Secp256K1PubKey{
		Key: bz,
	}

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshalling.
func (m *Secp256K1PubKey) MarshalAminoJSON() ([]byte, error) {
	return m.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshalling.
func (m *Secp256K1PubKey) UnmarshalAminoJSON(bz []byte) error {
	return m.UnmarshalAmino(bz)
}

func (m *Secp256K1PrivKey) Bytes() []byte {
	return m.Key.Bytes()
}

func (m *Secp256K1PrivKey) Sign(msg []byte) ([]byte, error) {
	return m.Key.Sign(msg)
}

func (m *Secp256K1PrivKey) PubKey() tmcrypto.PubKey {
	return &Secp256K1PubKey{Key: m.Key.PubKey().(secp256k1.PubKey)}
}

func (m *Secp256K1PrivKey) Equals(key tmcrypto.PrivKey) bool {
	return m.Key.Equals(key)
}

func (m *Secp256K1PrivKey) Type() string {
	return m.Key.Type()
}

// MarshalAminoJSON overrides Amino binary marshalling.
func (m Secp256K1PrivKey) MarshalAmino() ([]byte, error) {
	return m.Key.Bytes(), nil
}

// UnmarshalAminoJSON overrides Amino binary marshalling.
func (m *Secp256K1PrivKey) UnmarshalAmino(bz []byte) error {
	*m = Secp256K1PrivKey{
		Key: bz,
	}

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshalling.
func (m Secp256K1PrivKey) MarshalAminoJSON() ([]byte, error) {
	return m.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshalling.
func (m *Secp256K1PrivKey) UnmarshalAminoJSON(bz []byte) error {
	return m.UnmarshalAmino(bz)
}
