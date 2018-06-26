package crypto

import (
	"fmt"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	ledger "github.com/zondax/ledger-goclient"

	tcrypto "github.com/tendermint/tendermint/crypto"
)

func pubkeyLedgerSecp256k1(device *ledger.Ledger, path DerivationPath) (pub tcrypto.PubKey, err error) {
	key, err := device.GetPublicKeySECP256K1(path)
	if err != nil {
		return nil, fmt.Errorf("error fetching public key: %v", err)
	}
	var p tcrypto.PubKeySecp256k1
	// Reserialize in the 33-byte compressed format
	cmp, err := secp256k1.ParsePubKey(key[:], secp256k1.S256())
	copy(p[:], cmp.SerializeCompressed())
	pub = p
	return
}

func signLedgerSecp256k1(device *ledger.Ledger, path DerivationPath, msg []byte) (sig tcrypto.Signature, err error) {
	bsig, err := device.SignSECP256K1(path, msg)
	if err != nil {
		return sig, err
	}
	sig = tcrypto.SignatureSecp256k1FromBytes(bsig)
	return
}

// PrivKeyLedgerSecp256k1 implements PrivKey, calling the ledger nano
// we cache the PubKey from the first call to use it later
type PrivKeyLedgerSecp256k1 struct {
	// PubKey should be private, but we want to encode it via go-amino
	// so we can view the address later, even without having the ledger
	// attached
	CachedPubKey tcrypto.PubKey
	Path         DerivationPath
}

// NewPrivKeyLedgerSecp256k1 will generate a new key and store the
// public key for later use.
func NewPrivKeyLedgerSecp256k1(path DerivationPath) (tcrypto.PrivKey, error) {
	var pk PrivKeyLedgerSecp256k1
	pk.Path = path
	// cache the pubkey for later use
	pubKey, err := pk.getPubKey()
	if err != nil {
		return nil, err
	}
	pk.CachedPubKey = pubKey
	return &pk, err
}

// ValidateKey allows us to verify the sanity of a key
// after loading it from disk
func (pk PrivKeyLedgerSecp256k1) ValidateKey() error {
	// getPubKey will return an error if the ledger is not
	pub, err := pk.getPubKey()
	if err != nil {
		return err
	}
	// verify this matches cached address
	if !pub.Equals(pk.CachedPubKey) {
		return fmt.Errorf("cached key does not match retrieved key")
	}
	return nil
}

// AssertIsPrivKeyInner fulfils PrivKey Interface
func (pk *PrivKeyLedgerSecp256k1) AssertIsPrivKeyInner() {}

// Bytes fulfils PrivKey Interface - but it stores the cached pubkey so we can verify
// the same key when we reconnect to a ledger
func (pk PrivKeyLedgerSecp256k1) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(pk)
}

// Sign calls the ledger and stores the PubKey for future use
//
// Communication is checked on NewPrivKeyLedger and PrivKeyFromBytes,
// returning an error, so this should only trigger if the privkey is held
// in memory for a while before use.
func (pk PrivKeyLedgerSecp256k1) Sign(msg []byte) (tcrypto.Signature, error) {
	dev, err := getLedger()
	if err != nil {
		return nil, err
	}
	sig, err := signLedgerSecp256k1(dev, pk.Path, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// PubKey returns the stored PubKey
func (pk PrivKeyLedgerSecp256k1) PubKey() tcrypto.PubKey {
	return pk.CachedPubKey
}

// getPubKey reads the pubkey the ledger itself
// since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling
func (pk PrivKeyLedgerSecp256k1) getPubKey() (key tcrypto.PubKey, err error) {
	dev, err := getLedger()
	if err != nil {
		return key, fmt.Errorf("cannot connect to Ledger device - error: %v", err)
	}
	key, err = pubkeyLedgerSecp256k1(dev, pk.Path)
	if err != nil {
		return key, fmt.Errorf("please open Cosmos app on the Ledger device - error: %v", err)
	}
	return key, err
}

// Equals fulfils PrivKey Interface - makes sure both keys refer to the
// same
func (pk PrivKeyLedgerSecp256k1) Equals(other tcrypto.PrivKey) bool {
	if ledger, ok := other.(*PrivKeyLedgerSecp256k1); ok {
		return pk.CachedPubKey.Equals(ledger.CachedPubKey)
	}
	return false
}
