package crypto

import (
	"fmt"

	"github.com/pkg/errors"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
)

var (
	// discoverLedger defines a function to be invoked at runtime for discovering
	// a connected Ledger device.
	discoverLedger discoverLedgerFn
)

type (
	// discoverLedgerFn defines a Ledger discovery function that returns a
	// connected device or an error upon failure. Its allows a method to avoid CGO
	// dependencies when Ledger support is potentially not enabled.
	discoverLedgerFn func() (LedgerSECP256K1, error)

	// DerivationPath represents a Ledger derivation path.
	DerivationPath []uint32

	// LedgerSECP256K1 reflects an interface a Ledger API must implement for
	// the SECP256K1 scheme.
	LedgerSECP256K1 interface {
		GetPublicKeySECP256K1([]uint32) ([]byte, error)
		SignSECP256K1([]uint32, []byte) ([]byte, error)
	}

	// PrivKeyLedgerSecp256k1 implements PrivKey, calling the ledger nano we
	// cache the PubKey from the first call to use it later.
	PrivKeyLedgerSecp256k1 struct {
		// CachedPubKey should be private, but we want to encode it via
		// go-amino so we can view the address later, even without having the
		// ledger attached.
		CachedPubKey tmcrypto.PubKey
		Path         DerivationPath
		ledger       LedgerSECP256K1
	}
)

// NewPrivKeyLedgerSecp256k1 will generate a new key and store the public key
// for later use.
//
// CONTRACT: The ledger device, ledgerDevice, must be loaded and set prior to
// any creation of a PrivKeyLedgerSecp256k1.
func NewPrivKeyLedgerSecp256k1(path DerivationPath) (tmcrypto.PrivKey, error) {
	if discoverLedger == nil {
		return nil, errors.New("no Ledger discovery function defined")
	}

	device, err := discoverLedger()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create PrivKeyLedgerSecp256k1")
	}

	pkl := &PrivKeyLedgerSecp256k1{Path: path, ledger: device}

	pubKey, err := pkl.getPubKey()
	if err != nil {
		return nil, err
	}

	pkl.CachedPubKey = pubKey
	return pkl, err
}

// PubKey returns the cached public key.
func (pkl PrivKeyLedgerSecp256k1) PubKey() tmcrypto.PubKey {
	return pkl.CachedPubKey
}

// ValidateKey allows us to verify the sanity of a public key after loading it
// from disk.
func (pkl PrivKeyLedgerSecp256k1) ValidateKey() error {
	// getPubKey will return an error if the ledger is not
	pub, err := pkl.getPubKey()
	if err != nil {
		return err
	}

	// verify this matches cached address
	if !pub.Equals(pkl.CachedPubKey) {
		return fmt.Errorf("cached key does not match retrieved key")
	}

	return nil
}

// AssertIsPrivKeyInner implements the PrivKey interface. It performs a no-op.
func (pkl *PrivKeyLedgerSecp256k1) AssertIsPrivKeyInner() {}

// Bytes implements the PrivKey interface. It stores the cached public key so
// we can verify the same key when we reconnect to a ledger.
func (pkl PrivKeyLedgerSecp256k1) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(pkl)
}

// Equals implements the PrivKey interface. It makes sure two private keys
// refer to the same public key.
func (pkl PrivKeyLedgerSecp256k1) Equals(other tmcrypto.PrivKey) bool {
	if ledger, ok := other.(*PrivKeyLedgerSecp256k1); ok {
		return pkl.CachedPubKey.Equals(ledger.CachedPubKey)
	}

	return false
}

// Sign calls the ledger and stores the PubKey for future use.
//
// Communication is checked on NewPrivKeyLedger and PrivKeyFromBytes, returning
// an error, so this should only trigger if the private key is held in memory
// for a while before use.
func (pkl PrivKeyLedgerSecp256k1) Sign(msg []byte) ([]byte, error) {
	sig, err := pkl.signLedgerSecp256k1(msg)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// getPubKey reads the pubkey the ledger itself
// since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling
func (pkl PrivKeyLedgerSecp256k1) getPubKey() (key tmcrypto.PubKey, err error) {
	key, err = pkl.pubkeyLedgerSecp256k1()
	if err != nil {
		return key, fmt.Errorf("please open Cosmos app on the Ledger device - error: %v", err)
	}

	return key, err
}

func (pkl PrivKeyLedgerSecp256k1) signLedgerSecp256k1(msg []byte) ([]byte, error) {
	return pkl.ledger.SignSECP256K1(pkl.Path, msg)
}

func (pkl PrivKeyLedgerSecp256k1) pubkeyLedgerSecp256k1() (pub tmcrypto.PubKey, err error) {
	key, err := pkl.ledger.GetPublicKeySECP256K1(pkl.Path)
	if err != nil {
		return nil, fmt.Errorf("error fetching public key: %v", err)
	}

	var pk tmsecp256k1.PubKeySecp256k1

	// re-serialize in the 33-byte compressed format
	cmp, err := secp256k1.ParsePubKey(key[:], secp256k1.S256())
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}
	copy(pk[:], cmp.SerializeCompressed())

	return pk, nil
}
