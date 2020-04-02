package keyring

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/99designs/keyring"
	"github.com/pkg/errors"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/go-bip39"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

var (
	_ Keyring = &keystore{}
)

// NewKeyring creates a new instance of a keyring. Keybase
// options can be applied when generating this new Keybase.
// Available backends are "os", "file", "kwallet", "pass", "test".
func New(
	appName, backend, rootDir string, userInput io.Reader, opts ...AltKeyringOption,
) (Keyring, error) {

	var db keyring.Keyring
	var err error

	switch backend {
	case BackendMemory:
		return NewInMemory(opts...), err
	case BackendTest:
		db, err = keyring.Open(lkbToKeyringConfig(appName, rootDir, nil, true))
	case BackendFile:
		db, err = keyring.Open(newFileBackendKeyringConfig(appName, rootDir, userInput))
	case BackendOS:
		db, err = keyring.Open(lkbToKeyringConfig(appName, rootDir, userInput, false))
	case BackendKWallet:
		db, err = keyring.Open(newKWalletBackendKeyringConfig(appName, rootDir, userInput))
	case BackendPass:
		db, err = keyring.Open(newPassBackendKeyringConfig(appName, rootDir, userInput))
	default:
		return nil, fmt.Errorf("unknown keyring backend %v", backend)
	}

	if err != nil {
		return nil, err
	}

	return newKeystore(db, opts...), nil
}

type keystore struct {
	db      keyring.Keyring
	options altKrOptions
}

func newKeystore(kr keyring.Keyring, opts ...AltKeyringOption) keystore {
	// Default options for keybase
	options := altKrOptions{
		supportedAlgos:       AltSigningAlgoList{AltSecp256k1},
		supportedAlgosLedger: AltSigningAlgoList{AltSecp256k1},
	}

	for _, optionFn := range opts {
		optionFn(&options)
	}

	return keystore{kr, options}
}

func (ks keystore) ExportPubKeyArmor(uid string) (string, error) {
	bz, err := ks.Key(uid)
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name: %s", uid)
	}

	return crypto.ArmorPubKeyBytes(bz.GetPubKey().Bytes(), string(bz.GetAlgo())), nil
}

func (ks keystore) ExportPubKeyArmorByAddress(address types.Address) (string, error) {
	info, err := ks.KeyByAddress(address)
	if err != nil {
		return "", err
	}

	return ks.ExportPubKeyArmor(info.GetName())
}

func (ks keystore) ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error) {
	priv, err := ks.ExportPrivateKeyObject(uid)
	if err != nil {
		return "", err
	}

	info, err := ks.Key(uid)
	if err != nil {
		return "", err
	}

	return crypto.EncryptArmorPrivKey(priv, encryptPassphrase, string(info.GetAlgo())), nil
}

// ExportPrivateKeyObject exports an armored private key object.
func (ks keystore) ExportPrivateKeyObject(uid string) (tmcrypto.PrivKey, error) {
	info, err := ks.Key(uid)
	if err != nil {
		return nil, err
	}

	var priv tmcrypto.PrivKey

	switch linfo := info.(type) {
	case localInfo:
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return nil, err
		}

		priv, err = cryptoAmino.PrivKeyFromBytes([]byte(linfo.PrivKeyArmor))
		if err != nil {
			return nil, err
		}

	case ledgerInfo, offlineInfo, multiInfo:
		return nil, errors.New("only works on local private keys")
	}

	return priv, nil
}

func (ks keystore) ExportPrivKeyArmorByAddress(address types.Address, encryptPassphrase string) (armor string, err error) {
	byAddress, err := ks.KeyByAddress(address)
	if err != nil {
		return "", err
	}

	return ks.ExportPrivKeyArmor(byAddress.GetName(), encryptPassphrase)
}

func (ks keystore) ImportPrivKey(uid, armor, passphrase string) error {
	if ks.hasKey(uid) {
		return fmt.Errorf("cannot overwrite key: %s", uid)
	}

	privKey, algo, err := crypto.UnarmorDecryptPrivKey(armor, passphrase)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt private key")
	}

	_, err = ks.writeLocalKey(uid, privKey, pubKeyType(algo))
	if err != nil {
		return err
	}

	return nil
}

// HasKey returns whether the key exists in the keyring.
func (ks keystore) hasKey(name string) bool {
	bz, _ := ks.Key(name)
	return bz != nil
}

func (ks keystore) ImportPubKey(uid string, armor string) error {
	bz, _ := ks.Key(uid)
	if bz != nil {
		pubkey := bz.GetPubKey()

		if len(pubkey.Bytes()) > 0 {
			return fmt.Errorf("cannot overwrite data for name: %s", uid)
		}
	}

	pubBytes, algo, err := crypto.UnarmorPubKeyBytes(armor)
	if err != nil {
		return err
	}

	pubKey, err := cryptoAmino.PubKeyFromBytes(pubBytes)
	if err != nil {
		return err
	}

	_, err = ks.writeOfflineKey(uid, pubKey, pubKeyType(algo))
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) Sign(uid string, msg []byte) ([]byte, tmcrypto.PubKey, error) {
	info, err := ks.Key(uid)
	if err != nil {
		return nil, nil, err
	}

	var priv tmcrypto.PrivKey

	switch i := info.(type) {
	case localInfo:
		if i.PrivKeyArmor == "" {
			return nil, nil, fmt.Errorf("private key not available")
		}

		priv, err = cryptoAmino.PrivKeyFromBytes([]byte(i.PrivKeyArmor))
		if err != nil {
			return nil, nil, err
		}

	case ledgerInfo:
		return SignWithLedger(info, msg)

	case offlineInfo, multiInfo:
		return nil, info.GetPubKey(), errors.New("cannot sign with offline keys")
	}

	sig, err := priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, priv.PubKey(), nil
}

func (ks keystore) SignByAddress(address types.Address, msg []byte) ([]byte, tmcrypto.PubKey, error) {
	key, err := ks.KeyByAddress(address)
	if err != nil {
		return nil, nil, err
	}

	return ks.Sign(key.GetName(), msg)
}

func (ks keystore) SaveLedgerKey(uid string, algo AltSigningAlgo, hrp string, account, index uint32) (Info, error) {
	if !ks.options.supportedAlgosLedger.Contains(algo) {
		return nil, ErrUnsupportedSigningAlgo
	}

	coinType := types.GetConfig().GetCoinType()
	hdPath := hd.NewFundraiserParams(account, coinType, index)

	priv, _, err := crypto.NewPrivKeyLedgerSecp256k1(*hdPath, hrp)
	if err != nil {
		return nil, err
	}

	return ks.writeLedgerKey(uid, priv.PubKey(), *hdPath, algo.Name())
}

func (ks keystore) writeLedgerKey(name string, pub tmcrypto.PubKey, path hd.BIP44Params, algo pubKeyType) (Info, error) {
	info := newLedgerInfo(name, pub, path, algo)
	err := ks.writeInfo(name, info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

type altKrOptions struct {
	supportedAlgos       AltSigningAlgoList
	supportedAlgosLedger AltSigningAlgoList
}

func (ks keystore) SaveMultisig(uid string, pubkey tmcrypto.PubKey) (Info, error) {
	return ks.writeMultisigKey(uid, pubkey)
}

func (ks keystore) SavePubKey(uid string, pubkey tmcrypto.PubKey, algo pubKeyType) (Info, error) {
	return ks.writeOfflineKey(uid, pubkey, algo)
}

func (ks keystore) DeleteByAddress(address types.Address) error {
	info, err := ks.KeyByAddress(address)
	if err != nil {
		return err
	}

	err = ks.Delete(info.GetName())
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) Delete(uid string) error {
	info, err := ks.Key(uid)
	if err != nil {
		return err
	}

	err = ks.db.Remove(addrHexKeyAsString(info.GetAddress()))
	if err != nil {
		return err
	}

	err = ks.db.Remove(string(infoKey(uid)))
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) KeyByAddress(address types.Address) (Info, error) {
	ik, err := ks.db.Get(addrHexKeyAsString(address))
	if err != nil {
		return nil, err
	}

	if len(ik.Data) == 0 {
		return nil, fmt.Errorf("key with address %s not found", address)
	}

	bs, err := ks.db.Get(string(ik.Data))
	if err != nil {
		return nil, err
	}

	return unmarshalInfo(bs.Data)
}

func (ks keystore) List() ([]Info, error) {
	var res []Info
	keys, err := ks.db.Keys()
	if err != nil {
		return nil, err
	}

	sort.Strings(keys)

	for _, key := range keys {
		if strings.HasSuffix(key, infoSuffix) {
			rawInfo, err := ks.db.Get(key)
			if err != nil {
				return nil, err
			}

			if len(rawInfo.Data) == 0 {
				return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, key)
			}

			info, err := unmarshalInfo(rawInfo.Data)
			if err != nil {
				return nil, err
			}

			res = append(res, info)
		}
	}

	return res, nil
}

func (ks keystore) NewMnemonic(uid string, language Language, algo AltSigningAlgo) (Info, string, error) {
	if language != English {
		return nil, "", ErrUnsupportedLanguage
	}

	if !ks.isSupportedSigningAlgo(algo) {
		return nil, "", ErrUnsupportedSigningAlgo
	}

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

	info, err := ks.NewAccount(uid, mnemonic, DefaultBIP39Passphrase, types.GetConfig().GetFullFundraiserPath(), algo)
	if err != nil {
		return nil, "", err
	}

	return info, mnemonic, err
}

func (ks keystore) NewAccount(uid string, mnemonic string, bip39Passphrase string, hdPath string, algo AltSigningAlgo) (Info, error) {
	if !ks.isSupportedSigningAlgo(algo) {
		return nil, ErrUnsupportedSigningAlgo
	}

	// create master key and derive first key for keyring
	derivedPriv, err := algo.DeriveKey()(mnemonic, bip39Passphrase, hdPath)
	if err != nil {
		return nil, err
	}

	privKey := algo.PrivKeyGen()(derivedPriv)

	return ks.writeLocalKey(uid, privKey, algo.Name())
}

func (ks keystore) isSupportedSigningAlgo(algo AltSigningAlgo) bool {
	return ks.options.supportedAlgos.Contains(algo)
}

func (ks keystore) Key(uid string) (Info, error) {
	key := infoKey(uid)

	bs, err := ks.db.Get(string(key))
	if err != nil {
		return nil, err
	}

	if len(bs.Data) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, uid)
	}

	return unmarshalInfo(bs.Data)
}

func (ks keystore) writeLocalKey(name string, priv tmcrypto.PrivKey, algo pubKeyType) (Info, error) {
	// encrypt private key using keyring
	pub := priv.PubKey()

	info := newLocalInfo(name, pub, string(priv.Bytes()), algo)
	err := ks.writeInfo(name, info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (ks keystore) writeInfo(name string, info Info) error {
	// write the info by key
	key := infoKey(name)
	serializedInfo := marshalInfo(info)

	err := ks.db.Set(keyring.Item{
		Key:  string(key),
		Data: serializedInfo,
	})
	if err != nil {
		return err
	}

	err = ks.db.Set(keyring.Item{
		Key:  addrHexKeyAsString(info.GetAddress()),
		Data: key,
	})
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) writeOfflineKey(name string, pub tmcrypto.PubKey, algo pubKeyType) (Info, error) {
	info := newOfflineInfo(name, pub, algo)
	err := ks.writeInfo(name, info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (ks keystore) writeMultisigKey(name string, pub tmcrypto.PubKey) (Info, error) {
	info := NewMultiInfo(name, pub)
	err := ks.writeInfo(name, info)
	if err != nil {
		return nil, err
	}

	return info, nil
}
