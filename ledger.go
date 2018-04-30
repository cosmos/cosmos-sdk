package crypto

import (
	"github.com/pkg/errors"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	ledger "github.com/zondax/ledger-goclient"
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

func signLedger(device *ledger.Ledger, msg []byte) (pub PubKey, sig Signature, err error) {
	bsig, err := device.Sign(msg)
	if err != nil {
		return pub, sig, err
	}
	sig = SignatureSecp256k1FromBytes(bsig)
	key, err := device.GetPublicKey()
	if err != nil {
		return pub, sig, err
	}
	var p PubKeySecp256k1
	// Reserialize in the 33-byte compressed format
	cmp, err := secp256k1.ParsePubKey(key[:], secp256k1.S256())
	copy(p[:], cmp.SerializeCompressed())
	return p, sig, nil
}

// PrivKeyLedgerSecp256k1 implements PrivKey, calling the ledger nano
// we cache the PubKey from the first call to use it later
type PrivKeyLedgerSecp256k1 struct {
	// PubKey should be private, but we want to encode it via go-amino
	// so we can view the address later, even without having the ledger
	// attached
	CachedPubKey PubKey
}

// NewPrivKeyLedgerSecp256k1 will generate a new key and store the
// public key for later use.
func NewPrivKeyLedgerSecp256k1() (PrivKey, error) {
	var pk PrivKeyLedgerSecp256k1
	// getPubKey will cache the pubkey for later use,
	// this allows us to return an error early if the ledger
	// is not plugged in
	_, err := pk.getPubKey()
	return &pk, err
}

// ValidateKey allows us to verify the sanity of a key
// after loading it from disk
func (pk PrivKeyLedgerSecp256k1) ValidateKey() error {
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
func (pk *PrivKeyLedgerSecp256k1) AssertIsPrivKeyInner() {}

// Bytes fulfils PrivKey Interface - but it stores the cached pubkey so we can verify
// the same key when we reconnect to a ledger
func (pk PrivKeyLedgerSecp256k1) Bytes() []byte {
	bin, err := cdc.MarshalBinaryBare(pk)
	if err != nil {
		panic(err)
	}
	return bin
}

// Sign calls the ledger and stores the PubKey for future use
//
// XXX/TODO: panics if there is an error communicating with the ledger.
//
// Communication is checked on NewPrivKeyLedger and PrivKeyFromBytes,
// returning an error, so this should only trigger if the privkey is held
// in memory for a while before use.
func (pk PrivKeyLedgerSecp256k1) Sign(msg []byte) Signature {
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
	if pk.CachedPubKey == nil {
		pk.CachedPubKey = pub
	} else if !pk.CachedPubKey.Equals(pub) {
		panic("signed with a different key than stored")
	}
	return sig
}

// PubKey returns the stored PubKey
// TODO: query the ledger if not there, once it is not volatile
func (pk PrivKeyLedgerSecp256k1) PubKey() PubKey {
	key, err := pk.getPubKey()
	if err != nil {
		panic(err)
	}
	return key
}

// getPubKey reads the pubkey from cache or from the ledger itself
// since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling
func (pk PrivKeyLedgerSecp256k1) getPubKey() (key PubKey, err error) {
	// if we have no pubkey, set it
	if pk.CachedPubKey == nil {
		pk.CachedPubKey, err = pk.forceGetPubKey()
	}
	return pk.CachedPubKey, err
}

// forceGetPubKey is like getPubKey but ignores any cached key
// and ensures we get it from the ledger itself.
func (pk PrivKeyLedgerSecp256k1) forceGetPubKey() (key PubKey, err error) {
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
func (pk PrivKeyLedgerSecp256k1) Equals(other PrivKey) bool {
	if ledger, ok := other.(*PrivKeyLedgerSecp256k1); ok {
		return pk.CachedPubKey.Equals(ledger.CachedPubKey)
	}
	return false
}
