package cryptostore

import (
	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-keys"
)

// encryptedStorage needs passphrase to get private keys
type encryptedStorage struct {
	coder Encoder
	store keys.Storage
}

func (es encryptedStorage) Put(name, pass string, key crypto.PrivKey) error {
	secret, err := es.coder.Encrypt(key, pass)
	if err != nil {
		return err
	}

	ki := info(name, key)
	return es.store.Put(name, secret, ki)
}

func (es encryptedStorage) Get(name, pass string) (crypto.PrivKey, keys.KeyInfo, error) {
	secret, info, err := es.store.Get(name)
	if err != nil {
		return nil, info, err
	}
	key, err := es.coder.Decrypt(secret, pass)
	return key, info, err
}

func (es encryptedStorage) List() ([]keys.KeyInfo, error) {
	return es.store.List()
}

func (es encryptedStorage) Delete(name string) error {
	return es.store.Delete(name)
}

// info hardcodes the encoding of keys
func info(name string, key crypto.PrivKey) keys.KeyInfo {
	return keys.KeyInfo{
		Name:   name,
		PubKey: crypto.PubKeyS{key.PubKey()},
	}
}
