package crypto

import (
	"fmt"
	"github.com/pkg/errors"

	// "github.com/tendermint/ed25519"
	ledger "github.com/zondax/ledger-goclient"
)

func pubkeyLedgerEd25519(device *ledger.Ledger, path DerivationPath) (pub PubKey, err error) {
	key, err := device.GetPublicKeyED25519(path)
	if err != nil {
		return pub, fmt.Errorf("Error fetching public key: %v", err)
	}
	var p PubKeyLedgerEd25519
	copy(p[:], key[0:32])
	return p, err
}

func signLedgerEd25519(device *ledger.Ledger, path DerivationPath, msg []byte) (sig Signature, err error) {
	bsig, err := device.SignED25519(path, msg)
	if err != nil {
		return sig, err
	}
	sig = SignatureEd25519FromBytes(bsig)
	return sig, nil
}

// PrivKeyLedgerEd25519 implements PrivKey, calling the ledger nano
// we cache the PubKey from the first call to use it later
type PrivKeyLedgerEd25519 struct {
	// PubKey should be private, but we want to encode it via go-amino
	// so we can view the address later, even without having the ledger
	// attached
	CachedPubKey PubKey
	Path         DerivationPath
}

// NewPrivKeyLedgerEd25519 will generate a new key and store the
// public key for later use.
func NewPrivKeyLedgerEd25519(path DerivationPath) (PrivKey, error) {
	var pk PrivKeyLedgerEd25519
	pk.Path = path
	// getPubKey will cache the pubkey for later use,
	// this allows us to return an error early if the ledger
	// is not plugged in
	_, err := pk.getPubKey()
	return &pk, err
}

// ValidateKey allows us to verify the sanity of a key
// after loading it from disk
func (pk PrivKeyLedgerEd25519) ValidateKey() error {
	// getPubKey will return an error if the ledger is not
	// properly set up...
	pub, err := pk.forceGetPubKey()
	if err != nil {
		return err
	}
	// verify this matches cached address
	if !pub.Equals(pk.CachedPubKey) {
		return errors.New("Cached key does not match retrieved key")
	}
	return nil
}

// AssertIsPrivKeyInner fulfils PrivKey Interface
func (pk *PrivKeyLedgerEd25519) AssertIsPrivKeyInner() {}

// Bytes fulfils PrivKey Interface - but it stores the cached pubkey so we can verify
// the same key when we reconnect to a ledger
func (pk PrivKeyLedgerEd25519) Bytes() []byte {
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
func (pk PrivKeyLedgerEd25519) Sign(msg []byte) Signature {
	// oh, I wish there was better error handling
	dev, err := getLedger()
	if err != nil {
		panic(err)
	}

	sig, err := signLedgerEd25519(dev, pk.Path, msg)
	if err != nil {
		panic(err)
	}

	pub, err := pubkeyLedgerEd25519(dev, pk.Path)
	if err != nil {
		panic(err)
	}

	// if we have no pubkey yet, store it for future queries
	if pk.CachedPubKey == nil {
		pk.CachedPubKey = pub
	} else if !pk.CachedPubKey.Equals(pub) {
		panic("Stored key does not match signing key")
	}
	return sig
}

// PubKey returns the stored PubKey
// TODO: query the ledger if not there, once it is not volatile
func (pk PrivKeyLedgerEd25519) PubKey() PubKey {
	key, err := pk.getPubKey()
	if err != nil {
		panic(err)
	}
	return key
}

// getPubKey reads the pubkey from cache or from the ledger itself
// since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling
func (pk PrivKeyLedgerEd25519) getPubKey() (key PubKey, err error) {
	// if we have no pubkey, set it
	if pk.CachedPubKey == nil {
		pk.CachedPubKey, err = pk.forceGetPubKey()
	}
	return pk.CachedPubKey, err
}

// forceGetPubKey is like getPubKey but ignores any cached key
// and ensures we get it from the ledger itself.
func (pk PrivKeyLedgerEd25519) forceGetPubKey() (key PubKey, err error) {
	dev, err := getLedger()
	if err != nil {
		return key, errors.New(fmt.Sprintf("Cannot connect to Ledger device - error: %v", err))
	}
	key, err = pubkeyLedgerEd25519(dev, pk.Path)
	if err != nil {
		return key, errors.New(fmt.Sprintf("Please open Cosmos app on the Ledger device - error: %v", err))
	}
	return key, err
}

// Equals fulfils PrivKey Interface - makes sure both keys refer to the
// same
func (pk PrivKeyLedgerEd25519) Equals(other PrivKey) bool {
	if ledger, ok := other.(*PrivKeyLedgerEd25519); ok {
		return pk.CachedPubKey.Equals(ledger.CachedPubKey)
	}
	return false
}
