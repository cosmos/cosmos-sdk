package cryptostore

import (
	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-crypto/keys"
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

func (es encryptedStorage) Get(name, pass string) (crypto.PrivKey, keys.Info, error) {
	secret, info, err := es.store.Get(name)
	if err != nil {
		return crypto.PrivKey{}, info, err
	}
	key, err := es.coder.Decrypt(secret, pass)
	return key, info, err
}

func (es encryptedStorage) List() (keys.Infos, error) {
	return es.store.List()
}

func (es encryptedStorage) Delete(name string) error {
	return es.store.Delete(name)
}

// info hardcodes the encoding of keys
func info(name string, key crypto.PrivKey) keys.Info {
	pub := key.PubKey()
	return keys.Info{
		Name:    name,
		Address: pub.Address(),
		PubKey:  pub,
	}
}
