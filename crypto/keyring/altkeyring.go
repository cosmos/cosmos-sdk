package keyring

import (
	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

var (
	_ Keyring = &altKeyring{}
)

type altKeyring struct {
	db keyring.Keyring
}

func (a altKeyring) List() ([]Info, error) {
	panic("implement me")
}

func (a altKeyring) NewMnemonic(uid string, language Language, algo SigningAlgo) (Info, string, error) {
	if language != English {
		return nil, "", ErrUnsupportedLanguage
	}

	//if !IsSupportedAlgorithm(a.SupportedAlgos(), algo) {
	//	return nil, "", ErrUnsupportedSigningAlgo
	//}

	// Default number of words (24): This generates a mnemonic directly from the
	// number of words by reading system entropy.
	entropy, err := bip39.NewEntropy(defaultEntropySize)
	if err != nil {
		return nil, "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, "", err
	}

	info, err := a.NewAccount(uid, mnemonic, DefaultBIP39Passphrase, types.GetConfig().GetFullFundraiserPath(), algo)
	if err != nil {
		return nil, "", err
	}

	return info, mnemonic, err
}

func (a altKeyring) NewAccount(uid string, mnemonic string, bip39Passphrase string, hdPath string, algo SigningAlgo) (Info, error) {
	// create master key and derive first key for keyring
	derivedPriv, err := StdDeriveKey(mnemonic, bip39Passphrase, hdPath, algo)
	if err != nil {
		return nil, err
	}

	privKey, err := StdPrivKeyGen(derivedPriv, algo)
	if err != nil {
		return nil, err
	}

	var info Info

	info = a.writeLocalKey(uid, privKey, algo)

	return info, nil
}

func (a altKeyring) writeLocalKey(name string, priv tmcrypto.PrivKey, algo SigningAlgo) Info {
	// encrypt private key using keyring
	pub := priv.PubKey()

	info := newLocalInfo(name, pub, string(priv.Bytes()), algo)
	a.writeInfo(name, info)

	return info
}

func (a altKeyring) writeInfo(name string, info Info) {
	// write the info by key
	key := infoKey(name)
	serializedInfo := marshalInfo(info)

	err := a.db.Set(keyring.Item{
		Key:  string(key),
		Data: serializedInfo,
	})
	if err != nil {
		panic(err)
	}

	err = a.db.Set(keyring.Item{
		Key:  string(addrKey(info.GetAddress())),
		Data: key,
	})
	if err != nil {
		panic(err)
	}
}
