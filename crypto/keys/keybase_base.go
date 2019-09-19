package keys

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"

	tmcrypto "github.com/tendermint/tendermint/crypto"
)

// Auxiliary type that groups storage agnostic features together.

type baseKeybase struct{}

type persistDerivedKeyer interface {
	persistDerivedKey(seed []byte, passwd, name, fullHdPath string) (info Info, err error)
}

// SignWithLedger signs a binary message with the ledger device referenced by an Info object
// and returns the signed bytes and the public key. It returns an error if the device could
// not be queried or it returned an error.
func (kb baseKeybase) SignWithLedger(info Info, msg []byte) (sig []byte, pub tmcrypto.PubKey, err error) {
	i := info.(ledgerInfo)
	priv, err := crypto.NewPrivKeyLedgerSecp256k1Unsafe(i.Path)
	if err != nil {
		return
	}

	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, priv.PubKey(), nil
}

// DecodeSignature decodes a an length-prefixed binary signature from standard input
// and return it as a byte slice.
func (kb baseKeybase) DecodeSignature(info Info, msg []byte) (sig []byte, pub tmcrypto.PubKey, err error) {
	_, err = fmt.Fprintf(os.Stderr, "Message to sign:\n\n%s\n", msg)
	if err != nil {
		return nil, nil, err
	}

	buf := bufio.NewReader(os.Stdin)
	_, err = fmt.Fprintf(os.Stderr, "\nEnter Amino-encoded signature:\n")
	if err != nil {
		return nil, nil, err
	}

	// Will block until user inputs the signature
	signed, err := buf.ReadString('\n')
	if err != nil {
		return nil, nil, err
	}

	if err := cdc.UnmarshalBinaryLengthPrefixed([]byte(signed), sig); err != nil {
		return nil, nil, errors.Wrap(err, "failed to decode signature")
	}

	return sig, info.GetPubKey(), nil
}

// CreateLedger creates a new reference to a Ledger key pair.
// It returns a public key and a derivation path; it returns an error if the device
// could not be querier.
func (kb baseKeybase) CreateLedger(algo SigningAlgo, hrp string, account, index uint32) (tmcrypto.PubKey, *hd.BIP44Params, error) {
	if !kb.IsAlgoSupported(algo) {
		return nil, nil, ErrUnsupportedSigningAlgo
	}

	coinType := types.GetConfig().GetCoinType()
	hdPath := hd.NewFundraiserParams(account, coinType, index)
	priv, _, err := crypto.NewPrivKeyLedgerSecp256k1(*hdPath, hrp)
	if err != nil {
		return nil, nil, err
	}
	return priv.PubKey(), hdPath, nil
}

// CreateHDPath returns BIP 44 object from account and index parameters.
func (kb baseKeybase) CreateHDPath(account uint32, index uint32) *hd.BIP44Params {
	return hd.NewFundraiserParams(account, types.GetConfig().GetCoinType(), index)
}

// CreateMnemonic generates a new key with the given algorithm and language pair.
func (kb baseKeybase) CreateMnemonic(persistDerivedKeyer persistDerivedKeyer, name string,
	language Language, passwd string, algo SigningAlgo) (info Info, mnemonic string, err error) {

	if language != English {
		err = ErrUnsupportedLanguage
		return
	}
	if !kb.IsAlgoSupported(algo) {
		err = ErrUnsupportedSigningAlgo
		return
	}

	// default number of words (24):
	// this generates a mnemonic directly from the number of words by reading system entropy.
	entropy, err := bip39.NewEntropy(defaultEntropySize)
	if err != nil {
		return
	}
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return
	}

	info, err = persistDerivedKeyer.persistDerivedKey(
		bip39.NewSeed(mnemonic, DefaultBIP39Passphrase), passwd,
		name, types.GetConfig().GetFullFundraiserPath(),
	)
	return
}

// Derive computes a BIP39 seed from th mnemonic and bip39Passwd.
// Derive private key from the seed using the BIP44 params.
func (kb baseKeybase) Derive(persistDerivedKeyer persistDerivedKeyer, name, mnemonic,
	bip39Passphrase, encryptPasswd string, params hd.BIP44Params) (info Info, err error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return
	}

	info, err = persistDerivedKeyer.persistDerivedKey(seed, encryptPasswd, name, params.String())
	return
}

// ComputeDerivedKey derives and returns the private key for the given seed and HD path.
func (kb baseKeybase) ComputeDerivedKey(seed []byte, fullHdPath string) ([32]byte, error) {
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	return hd.DerivePrivateKeyForPath(masterPriv, ch, fullHdPath)
}

// IsAlgoSupported returns whether the signing algorithm is supported.
func (kb baseKeybase) IsAlgoSupported(algo SigningAlgo) bool { return algo == Secp256k1 }
