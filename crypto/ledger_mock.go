// +build ledger,test_ledger_mock

package crypto

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/pkg/errors"

	secp256k1 "github.com/tendermint/btcd/btcec"
	"github.com/tendermint/tendermint/crypto"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"

	bip39 "github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// GetPublicKeySECP256K1 mocks a ledger device
// as per the original API, it returns an uncompressed key
func (mock LedgerSECP256K1Mock) GetPublicKeySECP256K1(derivationPath []uint32) ([]byte, error) {
	if derivationPath[0] != 44 {
		return nil, errors.New("Invalid derivation path")
	}

	if derivationPath[1] != sdk.GetConfig().GetCoinType() {
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

// GetAddressPubKeySECP256K1 mocks a ledger device
// as per the original API, it returns a compressed key and a bech32 address
func (mock LedgerSECP256K1Mock) GetAddressPubKeySECP256K1(derivationPath []uint32, hrp string) ([]byte, string, error) {
	pk, err := mock.GetPublicKeySECP256K1(derivationPath)
	if err != nil {
		return nil, "", err
	}

	// re-serialize in the 33-byte compressed format
	cmp, err := btcec.ParsePubKey(pk[:], btcec.S256())
	if err != nil {
		return nil, "", fmt.Errorf("error parsing public key: %v", err)
	}

	var compressedPublicKey tmsecp256k1.PubKeySecp256k1
	copy(compressedPublicKey[:], cmp.SerializeCompressed())

	// Generate the bech32 addr using existing tmcrypto/etc.
	addr := sdk.AccAddress(compressedPublicKey.Address()).String()
	return pk, addr, err
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

// ShowAddressSECP256K1 shows the address for the corresponding bip32 derivation path
func (mock LedgerSECP256K1Mock) ShowAddressSECP256K1(bip32Path []uint32, hrp string) error {
	fmt.Printf("Request to show address for %v at %v", hrp, bip32Path)
	return nil
}
