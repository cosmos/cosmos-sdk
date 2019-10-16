package keys

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/types"
)

type (
	// baseKeybase is an auxiliary type that groups Keybase storage agnostic features
	// together.
	baseKeybase struct{}

	keyWriter interface {
		writeLocalKeyer
		infoWriter
	}

	writeLocalKeyer interface {
		writeLocalKey(name string, priv tmcrypto.PrivKey, passphrase string) Info
	}

	infoWriter interface {
		writeInfo(name string, info Info)
	}
)

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

	// will block until user inputs the signature
	signed, err := buf.ReadString('\n')
	if err != nil {
		return nil, nil, err
	}

	if err := cdc.UnmarshalBinaryLengthPrefixed([]byte(signed), sig); err != nil {
		return nil, nil, errors.Wrap(err, "failed to decode signature")
	}

	return sig, info.GetPubKey(), nil
}

// CreateAccount creates an account Info object.
func (kb baseKeybase) CreateAccount(
	keyWriter keyWriter, name, mnemonic, bip39Passwd, encryptPasswd string, account, index uint32,
) (Info, error) {

	hdPath := CreateHDPath(account, index)
	return kb.Derive(keyWriter, name, mnemonic, bip39Passwd, encryptPasswd, *hdPath)
}

func (kb baseKeybase) persistDerivedKey(
	keyWriter keyWriter, seed []byte, passwd, name, fullHdPath string,
) (Info, error) {

	// create master key and derive first key for keyring
	derivedPriv, err := ComputeDerivedKey(seed, fullHdPath)
	if err != nil {
		return nil, err
	}

	var info Info

	if passwd != "" {
		info = keyWriter.writeLocalKey(name, secp256k1.PrivKeySecp256k1(derivedPriv), passwd)
	} else {
		info = kb.writeOfflineKey(keyWriter, name, secp256k1.PrivKeySecp256k1(derivedPriv).PubKey())
	}

	return info, nil
}

// CreateLedger creates a new reference to a Ledger key pair. It returns a public
// key and a derivation path. It returns an error if the device could not be queried.
func (kb baseKeybase) CreateLedger(
	w infoWriter, name string, algo SigningAlgo, hrp string, account, index uint32,
) (Info, error) {

	if !IsAlgoSupported(algo) {
		return nil, ErrUnsupportedSigningAlgo
	}

	coinType := types.GetConfig().GetCoinType()
	hdPath := hd.NewFundraiserParams(account, coinType, index)

	priv, _, err := crypto.NewPrivKeyLedgerSecp256k1(*hdPath, hrp)
	if err != nil {
		return nil, err
	}

	return kb.writeLedgerKey(w, name, priv.PubKey(), *hdPath), nil
}

// CreateMnemonic generates a new key with the given algorithm and language pair.
func (kb baseKeybase) CreateMnemonic(
	keyWriter keyWriter, name string, language Language, passwd string, algo SigningAlgo,
) (info Info, mnemonic string, err error) {

	if language != English {
		return nil, "", ErrUnsupportedLanguage
	}

	if !IsAlgoSupported(algo) {
		return nil, "", ErrUnsupportedSigningAlgo
	}

	// Default number of words (24): This generates a mnemonic directly from the
	// number of words by reading system entropy.
	entropy, err := bip39.NewEntropy(defaultEntropySize)
	if err != nil {
		return nil, "", err
	}

	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, "", err
	}

	info, err = kb.persistDerivedKey(
		keyWriter,
		bip39.NewSeed(mnemonic, DefaultBIP39Passphrase), passwd,
		name, types.GetConfig().GetFullFundraiserPath(),
	)

	return info, mnemonic, err
}

// Derive computes a BIP39 seed from the mnemonic and bip39Passphrase. It creates
// a private key from the seed using the BIP44 params.
func (kb baseKeybase) Derive(
	keyWriter keyWriter, name, mnemonic, bip39Passphrase, encryptPasswd string, params hd.BIP44Params,
) (Info, error) {

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return nil, err
	}

	return kb.persistDerivedKey(keyWriter, seed, encryptPasswd, name, params.String())
}

func (kb baseKeybase) writeLedgerKey(w infoWriter, name string, pub tmcrypto.PubKey, path hd.BIP44Params) Info {
	info := newLedgerInfo(name, pub, path)
	w.writeInfo(name, info)
	return info
}

func (kb baseKeybase) writeOfflineKey(w infoWriter, name string, pub tmcrypto.PubKey) Info {
	info := newOfflineInfo(name, pub)
	w.writeInfo(name, info)
	return info
}

func (kb baseKeybase) writeMultisigKey(w infoWriter, name string, pub tmcrypto.PubKey) Info {
	info := NewMultiInfo(name, pub)
	w.writeInfo(name, info)
	return info
}

// ComputeDerivedKey derives and returns the private key for the given seed and HD path.
func ComputeDerivedKey(seed []byte, fullHdPath string) ([32]byte, error) {
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	return hd.DerivePrivateKeyForPath(masterPriv, ch, fullHdPath)
}

// CreateHDPath returns BIP 44 object from account and index parameters.
func CreateHDPath(account uint32, index uint32) *hd.BIP44Params {
	return hd.NewFundraiserParams(account, types.GetConfig().GetCoinType(), index)
}

// IsAlgoSupported returns whether the signing algorithm is supported.
//
// TODO: Refactor this to be configurable to support interchangeable key signing
// and addressing.
// Ref: https://github.com/cosmos/cosmos-sdk/issues/4941
func IsAlgoSupported(algo SigningAlgo) bool { return algo == Secp256k1 }
