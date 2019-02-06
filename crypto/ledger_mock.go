// +build ledger,test_ledger_mock

package crypto

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	secp256k1 "github.com/tendermint/btcd/btcec"
	"github.com/tendermint/tendermint/crypto"
)

// If ledger support (build tag) has been enabled, which implies a CGO dependency,
// set the discoverLedger function which is responsible for loading the Ledger
// device at runtime or returning an error.
func init() {
	discoverLedger = func() (LedgerSECP256K1, error) {
		return LedgerSECP256K1Mock{}, nil
	}
}

type LedgerSECP256K1Mock struct {
}

func (mock LedgerSECP256K1Mock) Close() error {
	return nil
}

func (mock LedgerSECP256K1Mock) GetPublicKeySECP256K1(derivationPath []uint32) ([]byte, error) {
	if derivationPath[0] != 44 {
		return nil, errors.New("Invalid derivation path")
	}
	if derivationPath[1] != 118 {
		return nil, errors.New("Invalid derivation path")
	}

	seed, err := bip39.NewSeedWithErrorChecking(tests.TestMnemonic, "")
	if err != nil {
		return nil, err
	}

	path := hd.NewParams(derivationPath[0], derivationPath[1], derivationPath[2], derivationPath[3] != 0, derivationPath[4])
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, path.String())
	if err != nil {
		return nil, err
	}

	_, pubkeyObject := secp256k1.PrivKeyFromBytes(secp256k1.S256(), derivedPriv[:])

	return pubkeyObject.SerializeUncompressed(), nil
}

func (mock LedgerSECP256K1Mock) SignSECP256K1(derivationPath []uint32, message []byte) ([]byte, error) {
	path := hd.NewParams(derivationPath[0], derivationPath[1], derivationPath[2], derivationPath[3] != 0, derivationPath[4])
	seed, err := bip39.NewSeedWithErrorChecking(tests.TestMnemonic, "")
	if err != nil {
		return nil, err
	}

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, path.String())
	if err != nil {
		return nil, err
	}

	priv, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), derivedPriv[:])

	sig, err := priv.Sign(crypto.Sha256(message))
	if err != nil {
		return nil, err
	}

	// Need to return DER as the ledger does
	sig2 := btcec.Signature{R: sig.R, S: sig.S}
	return sig2.Serialize(), nil
}
