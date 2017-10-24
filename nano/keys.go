package nano

import (
	"bytes"
	"encoding/hex"

	"github.com/pkg/errors"

	ledger "github.com/ethanfrey/ledger"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

//nolint
const (
	NameLedgerEd25519 = "ledger-ed25519"
	TypeLedgerEd25519 = 0x10

	// Timeout is the number of seconds to wait for a response from the ledger
	// if eg. waiting for user confirmation on button push
	Timeout = 20
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
		resp, err = device.Exchange(pack, Timeout)
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
	return PubKeyLedgerEd25519FromBytes(b), crypto.SignatureEd25519FromBytes(bsig), nil
}

// PrivKeyLedgerEd25519 implements PrivKey, calling the ledger nano
// we cache the PubKey from the first call to use it later
type PrivKeyLedgerEd25519 struct {
	// PubKey should be private, but we want to encode it via go-wire
	// so we can view the address later, even without having the ledger
	// attached
	CachedPubKey crypto.PubKey
}

// NewPrivKeyLedgerEd25519Ed25519 will generate a new key and store the
// public key for later use.
func NewPrivKeyLedgerEd25519Ed25519() (crypto.PrivKey, error) {
	var pk PrivKeyLedgerEd25519
	// getPubKey will cache the pubkey for later use,
	// this allows us to return an error early if the ledger
	// is not plugged in
	_, err := pk.getPubKey()
	return pk.Wrap(), err
}

// ValidateKey allows us to verify the sanity of a key
// after loading it from disk
func (pk *PrivKeyLedgerEd25519) ValidateKey() error {
	// getPubKey will return an error if the ledger is not
	// properly set up...
	pub, err := pk.forceGetPubKey()
	if err != nil {
		return err
	}
	// verify this matches cached address
	if !pub.Equals(pk.CachedPubKey) {
		return errors.New("ledger doesn't match cached key")
	}
	return nil
}

// AssertIsPrivKeyInner fulfils PrivKey Interface
func (pk *PrivKeyLedgerEd25519) AssertIsPrivKeyInner() {}

// Bytes fulfils pk Interface - stores the cached pubkey so we can verify
// the same key when we reconnect to a ledger
func (pk *PrivKeyLedgerEd25519) Bytes() []byte {
	return wire.BinaryBytes(pk.Wrap())
}

// Sign calls the ledger and stores the pk for future use
//
// XXX/TODO: panics if there is an error communicating with the ledger.
//
// Communication is checked on NewPrivKeyLedger and PrivKeyFromBytes,
// returning an error, so this should only trigger if the privkey is held
// in memory for a while before use.
func (pk *PrivKeyLedgerEd25519) Sign(msg []byte) crypto.Signature {
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
	if pk.CachedPubKey.Empty() {
		pk.CachedPubKey = pub
	} else if !pk.CachedPubKey.Equals(pub) {
		panic("signed with a different key than stored")
	}
	return sig
}

// PubKey returns the stored PubKey
// TODO: query the ledger if not there, once it is not volatile
func (pk *PrivKeyLedgerEd25519) PubKey() crypto.PubKey {
	key, err := pk.getPubKey()
	if err != nil {
		panic(err)
	}
	return key
}

// getPubKey reads the pubkey from cache or from the ledger itself
// since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling
func (pk *PrivKeyLedgerEd25519) getPubKey() (key crypto.PubKey, err error) {
	// if we have no pubkey, set it
	if pk.CachedPubKey.Empty() {
		pk.CachedPubKey, err = pk.forceGetPubKey()
	}
	return pk.CachedPubKey, err
}

// forceGetPubKey is like getPubKey but ignores any cached key
// and ensures we get it from the ledger itself.
func (pk *PrivKeyLedgerEd25519) forceGetPubKey() (key crypto.PubKey, err error) {
	dev, err := getLedger()
	if err != nil {
		return key, errors.New("Can't connect to ledger device")
	}
	key, _, err = signLedger(dev, []byte{0})
	if err != nil {
		return key, errors.New("Please open cosmos app on the ledger")
	}
	return key, err
}

// Equals fulfils PrivKey Interface - makes sure both keys refer to the
// same
func (pk *PrivKeyLedgerEd25519) Equals(other crypto.PrivKey) bool {
	if ledger, ok := other.Unwrap().(*PrivKeyLedgerEd25519); ok {
		return pk.CachedPubKey.Equals(ledger.CachedPubKey)
	}
	return false
}

// MockPrivKeyLedgerEd25519 behaves as the ledger, but stores a pre-packaged call-response
// for use in test cases
type MockPrivKeyLedgerEd25519 struct {
	Msg []byte
	Pub [KeyLength]byte
	Sig [SigLength]byte
}

// NewMockKey returns
func NewMockKey(msg, pubkey, sig string) (pk MockPrivKeyLedgerEd25519) {
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

var _ crypto.PrivKeyInner = MockPrivKeyLedgerEd25519{}

// AssertIsPrivKeyInner fulfils PrivKey Interface
func (pk MockPrivKeyLedgerEd25519) AssertIsPrivKeyInner() {}

// Bytes fulfils PrivKey Interface - not supported
func (pk MockPrivKeyLedgerEd25519) Bytes() []byte {
	return nil
}

// Sign returns a real SignatureLedger, if the msg matches what we expect
func (pk MockPrivKeyLedgerEd25519) Sign(msg []byte) crypto.Signature {
	if !bytes.Equal(pk.Msg, msg) {
		panic("Mock key is for different msg")
	}
	return crypto.SignatureEd25519(pk.Sig).Wrap()
}

// PubKey returns a real PubKeyLedgerEd25519, that will verify this signature
func (pk MockPrivKeyLedgerEd25519) PubKey() crypto.PubKey {
	return PubKeyLedgerEd25519FromBytes(pk.Pub)
}

// Equals compares that two Mocks have the same data
func (pk MockPrivKeyLedgerEd25519) Equals(other crypto.PrivKey) bool {
	if mock, ok := other.Unwrap().(MockPrivKeyLedgerEd25519); ok {
		return bytes.Equal(mock.Pub[:], pk.Pub[:]) &&
			bytes.Equal(mock.Sig[:], pk.Sig[:]) &&
			bytes.Equal(mock.Msg, pk.Msg)
	}
	return false
}

////////////////////////////////////////////
// pubkey

// PubKeyLedgerEd25519 works like a normal Ed25519 except a hash before the verify bytes
type PubKeyLedgerEd25519 struct {
	crypto.PubKeyEd25519
}

// PubKeyLedgerEd25519FromBytes creates a PubKey from the raw bytes
func PubKeyLedgerEd25519FromBytes(key [32]byte) crypto.PubKey {
	return PubKeyLedgerEd25519{crypto.PubKeyEd25519(key)}.Wrap()
}

// Bytes fulfils pk Interface - no data, just type info
func (pk PubKeyLedgerEd25519) Bytes() []byte {
	return wire.BinaryBytes(pk.Wrap())
}

// VerifyBytes uses the normal Ed25519 algorithm but a sha512 hash beforehand
func (pk PubKeyLedgerEd25519) VerifyBytes(msg []byte, sig crypto.Signature) bool {
	hmsg := hashMsg(msg)
	return pk.PubKeyEd25519.VerifyBytes(hmsg, sig)
}

// Equals implements PubKey interface
func (pk PubKeyLedgerEd25519) Equals(other crypto.PubKey) bool {
	if ledger, ok := other.Unwrap().(PubKeyLedgerEd25519); ok {
		return pk.PubKeyEd25519.Equals(ledger.PubKeyEd25519.Wrap())
	}
	return false
}

/*** registration with go-data ***/

func init() {
	crypto.PrivKeyMapper.
		RegisterImplementation(&PrivKeyLedgerEd25519{}, NameLedgerEd25519, TypeLedgerEd25519).
		RegisterImplementation(MockPrivKeyLedgerEd25519{}, "mock-ledger", 0x11)

	crypto.PubKeyMapper.
		RegisterImplementation(PubKeyLedgerEd25519{}, NameLedgerEd25519, TypeLedgerEd25519)
}

// Wrap fulfils interface for PrivKey struct
func (pk *PrivKeyLedgerEd25519) Wrap() crypto.PrivKey {
	return crypto.PrivKey{PrivKeyInner: pk}
}

// Wrap fulfils interface for PrivKey struct
func (pk MockPrivKeyLedgerEd25519) Wrap() crypto.PrivKey {
	return crypto.PrivKey{PrivKeyInner: pk}
}

// Wrap fulfils interface for PubKey struct
func (pk PubKeyLedgerEd25519) Wrap() crypto.PubKey {
	return crypto.PubKey{PubKeyInner: pk}
}
