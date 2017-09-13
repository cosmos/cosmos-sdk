package nano

import (
	"bytes"
	"encoding/hex"

	"github.com/pkg/errors"

	ledger "github.com/ethanfrey/ledger"

	crypto "github.com/tendermint/go-crypto"
)

var device *ledger.Ledger

// getLedger gets a copy of the device, and caches it
func getLedger() (*ledger.Ledger, error) {
	var err error
	if device == nil {
		device, err = ledger.FindLedger()
	}
	return device, err
}

func signLedger(device *ledger.Ledger, msg []byte) (pk crypto.PubKey, sig crypto.Signature, err error) {
	var resp []byte

	packets := generateSignRequests(msg)
	for _, pack := range packets {
		resp, err = device.Exchange(pack, 100)
		if err != nil {
			return pk, sig, err
		}
	}

	// the last call is the result we want and needs to be parsed
	key, bsig, err := parseDigest(resp)
	if err != nil {
		return pk, sig, err
	}

	var b [32]byte
	copy(b[:], key)
	return PubKeyLedgerFromBytes(b), crypto.SignatureEd25519FromBytes(bsig), nil
}

// PrivKeyLedger implements PrivKey, calling the ledger nano
// we cache the PubKey from the first call to use it later
type PrivKeyLedger struct {
	pubKey crypto.PubKey
}

func NewPrivKeyLedger() (crypto.PrivKey, error) {
	var pk PrivKeyLedger
	// getPubKey will cache the pubkey for later use,
	// this allows us to return an error early if the ledger
	// is not plugged in
	_, err := pk.getPubKey()
	return pk.Wrap(), err
}

// AssertIsPrivKeyInner fulfils PrivKey Interface
func (pk *PrivKeyLedger) AssertIsPrivKeyInner() {}

// Bytes fulfils pk Interface - not supported
func (pk *PrivKeyLedger) Bytes() []byte {
	return nil
}

// Sign calls the ledger and stores the pk for future use
func (pk *PrivKeyLedger) Sign(msg []byte) crypto.Signature {
	// oh, I wish there was better error handling
	dev, err := getLedger()
	if err != nil {
		panic(err)
	}

	pub, sig, err := signLedger(dev, msg)
	if err != nil {
		panic(err)
	}

	// if we have no pubkey yet, store it for future queries
	if pk.pubKey.Empty() {
		pk.pubKey = pub
	}
	return sig
}

// PubKey returns the stored PubKey
// TODO: query the ledger if not there, once it is not volatile
func (pk *PrivKeyLedger) PubKey() crypto.PubKey {
	key, err := pk.getPubKey()
	if err != nil {
		panic(err)
	}
	return key
}

// getPubKey reads the pubkey from cache or from the ledger itself
// since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling
func (pk *PrivKeyLedger) getPubKey() (key crypto.PubKey, err error) {
	// if we have no pubkey, set it
	if pk.pubKey.Empty() {
		dev, err := getLedger()
		if err != nil {
			return key, errors.WithMessage(err, "Can't connect to ledger")
		}
		pk.pubKey, _, err = signLedger(dev, []byte{0})
		if err != nil {
			return key, errors.WithMessage(err, "Can't sign with app")
		}
	}
	return pk.pubKey, nil
}

// Equals fulfils PrivKey Interface
// TODO: needs to be fixed
func (pk *PrivKeyLedger) Equals(other crypto.PrivKey) bool {
	if _, ok := other.Unwrap().(*PrivKeyLedger); ok {
		return true
	}
	return false
}

// MockPrivKeyLedger behaves as the ledger, but stores a pre-packaged call-response
// for use in test cases
type MockPrivKeyLedger struct {
	Msg []byte
	Pub [KeyLength]byte
	Sig [SigLength]byte
}

// NewMockKey returns
func NewMockKey(msg, pubkey, sig string) (pk MockPrivKeyLedger) {
	var err error
	pk.Msg, err = hex.DecodeString(msg)
	if err != nil {
		panic(err)
	}

	bpk, err := hex.DecodeString(pubkey)
	if err != nil {
		panic(err)
	}
	bsig, err := hex.DecodeString(sig)
	if err != nil {
		panic(err)
	}

	copy(pk.Pub[:], bpk)
	copy(pk.Sig[:], bsig)
	return pk
}

var _ crypto.PrivKeyInner = MockPrivKeyLedger{}

// AssertIsPrivKeyInner fulfils PrivKey Interface
func (pk MockPrivKeyLedger) AssertIsPrivKeyInner() {}

// Bytes fulfils PrivKey Interface - not supported
func (pk MockPrivKeyLedger) Bytes() []byte {
	return nil
}

// Sign returns a real SignatureLedger, if the msg matches what we expect
func (pk MockPrivKeyLedger) Sign(msg []byte) crypto.Signature {
	if !bytes.Equal(pk.Msg, msg) {
		panic("Mock key is for different msg")
	}
	return crypto.SignatureEd25519(pk.Sig).Wrap()
}

// PubKey returns a real PubKeyLedger, that will verify this signature
func (pk MockPrivKeyLedger) PubKey() crypto.PubKey {
	return PubKeyLedgerFromBytes(pk.Pub)
}

// Equals compares that two Mocks have the same data
func (pk MockPrivKeyLedger) Equals(other crypto.PrivKey) bool {
	if mock, ok := other.Unwrap().(MockPrivKeyLedger); ok {
		return bytes.Equal(mock.Pub[:], pk.Pub[:]) &&
			bytes.Equal(mock.Sig[:], pk.Sig[:]) &&
			bytes.Equal(mock.Msg, pk.Msg)
	}
	return false
}

////////////////////////////////////////////
// pubkey

// PubKeyLedger works like a normal Ed25519 except a hash before the verify bytes
type PubKeyLedger struct {
	crypto.PubKeyEd25519
}

// PubKeyLedgerFromBytes creates a PubKey from the raw bytes
func PubKeyLedgerFromBytes(key [32]byte) crypto.PubKey {
	return PubKeyLedger{crypto.PubKeyEd25519(key)}.Wrap()
}

// VerifyBytes uses the normal Ed25519 algorithm but a sha512 hash beforehand
func (pk PubKeyLedger) VerifyBytes(msg []byte, sig crypto.Signature) bool {
	hmsg := hashMsg(msg)
	return pk.PubKeyEd25519.VerifyBytes(hmsg, sig)
}

// Equals implements PubKey interface
func (pk PubKeyLedger) Equals(other crypto.PubKey) bool {
	if ledger, ok := other.Unwrap().(PubKeyLedger); ok {
		return bytes.Equal(pk.PubKeyEd25519[:], ledger.PubKeyEd25519[:])
	}
	return false
}

/*** registration with go-data ***/

func init() {
	crypto.PrivKeyMapper.
		RegisterImplementation(&PrivKeyLedger{}, "ledger", 0x10).
		RegisterImplementation(MockPrivKeyLedger{}, "mock-ledger", 0x11)

	crypto.PubKeyMapper.
		RegisterImplementation(PubKeyLedger{}, "ledger", 0x10)
}

// Wrap fulfils interface for PrivKey struct
func (pk *PrivKeyLedger) Wrap() crypto.PrivKey {
	return crypto.PrivKey{pk}
}

// Wrap fulfils interface for PrivKey struct
func (pk MockPrivKeyLedger) Wrap() crypto.PrivKey {
	return crypto.PrivKey{pk}
}

// Wrap fulfils interface for PubKey struct
func (pk PubKeyLedger) Wrap() crypto.PubKey {
	return crypto.PubKey{pk}
}
