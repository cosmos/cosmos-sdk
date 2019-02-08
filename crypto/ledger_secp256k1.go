package crypto

import (
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/pkg/errors"
	tmbtcec "github.com/tendermint/btcd/btcec"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
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

	// LedgerSECP256K1 reflects an interface a Ledger API must implement for
	// the SECP256K1 scheme.
	LedgerSECP256K1 interface {
		Close() error
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
		Path         hd.BIP44Params
	}
)

// NewPrivKeyLedgerSecp256k1 will generate a new key and store the public key
// for later use.
func NewPrivKeyLedgerSecp256k1(path hd.BIP44Params) (tmcrypto.PrivKey, error) {
	device, err := getLedgerDevice()
	if err != nil {
		return nil, err
	}
	defer warnIfErrors(device.Close)

	pubKey, err := getPubKey(device, path)
	if err != nil {
		return nil, err
	}

	return PrivKeyLedgerSecp256k1{pubKey, path}, nil
}

// PubKey returns the cached public key.
func (pkl PrivKeyLedgerSecp256k1) PubKey() tmcrypto.PubKey {
	return pkl.CachedPubKey
}

// Sign returns a secp256k1 signature for the corresponding message
func (pkl PrivKeyLedgerSecp256k1) Sign(message []byte) ([]byte, error) {
	device, err := getLedgerDevice()
	if err != nil {
		return nil, err
	}
	defer warnIfErrors(device.Close)

	return sign(device, pkl, message)
}

// ValidateKey allows us to verify the sanity of a public key after loading it
// from disk.
func (pkl PrivKeyLedgerSecp256k1) ValidateKey() error {
	device, err := getLedgerDevice()
	if err != nil {
		return err
	}
	defer warnIfErrors(device.Close)

	return validateKey(device, pkl)
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
	if otherKey, ok := other.(PrivKeyLedgerSecp256k1); ok {
		return pkl.CachedPubKey.Equals(otherKey.CachedPubKey)
	}
	return false
}

// warnIfErrors wraps a function and writes a warning to stderr. This is required
// to avoid ignoring errors when defer is used. Using defer may result in linter warnings.
func warnIfErrors(f func() error) {
	if err := f(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, "received error when closing ledger connection", err)
	}
}

func convertDERtoBER(signatureDER []byte) ([]byte, error) {
	sigDER, err := btcec.ParseDERSignature(signatureDER[:], btcec.S256())
	if err != nil {
		return nil, err
	}
	sigBER := tmbtcec.Signature{R: sigDER.R, S: sigDER.S}
	return sigBER.Serialize(), nil
}

func getLedgerDevice() (LedgerSECP256K1, error) {
	if discoverLedger == nil {
		return nil, errors.New("no Ledger discovery function defined")
	}

	device, err := discoverLedger()
	if err != nil {
		return nil, errors.Wrap(err, "ledger nano S")
	}

	return device, nil
}

func validateKey(device LedgerSECP256K1, pkl PrivKeyLedgerSecp256k1) error {
	pub, err := getPubKey(device, pkl.Path)
	if err != nil {
		return err
	}

	// verify this matches cached address
	if !pub.Equals(pkl.CachedPubKey) {
		return fmt.Errorf("cached key does not match retrieved key")
	}

	return nil
}

// Sign calls the ledger and stores the PubKey for future use.
//
// Communication is checked on NewPrivKeyLedger and PrivKeyFromBytes, returning
// an error, so this should only trigger if the private key is held in memory
// for a while before use.
func sign(device LedgerSECP256K1, pkl PrivKeyLedgerSecp256k1, msg []byte) ([]byte, error) {
	err := validateKey(device, pkl)
	if err != nil {
		return nil, err
	}

	sig, err := device.SignSECP256K1(pkl.Path.DerivationPath(), msg)
	if err != nil {
		return nil, err
	}

	return convertDERtoBER(sig)
}

// getPubKey reads the pubkey the ledger itself
// since this involves IO, it may return an error, which is not exposed
// in the PubKey interface, so this function allows better error handling
func getPubKey(device LedgerSECP256K1, path hd.BIP44Params) (tmcrypto.PubKey, error) {
	publicKey, err := device.GetPublicKeySECP256K1(path.DerivationPath())
	if err != nil {
		return nil, fmt.Errorf("please open Cosmos app on the Ledger device - error: %v", err)
	}

	// re-serialize in the 33-byte compressed format
	cmp, err := btcec.ParsePubKey(publicKey[:], btcec.S256())
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}

	var compressedPublicKey tmsecp256k1.PubKeySecp256k1
	copy(compressedPublicKey[:], cmp.SerializeCompressed())

	return compressedPublicKey, nil
}
