package keyring

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ Keyring = &altKeyring{}
)

// Keyring exposes operations on a generic keystore
type Keyring interface {
	// List all keys.
	List() ([]Info, error)

	// Key and KeyByAddress return keys by uid and address respectively.
	Key(uid string) (Info, error)
	KeyByAddress(address types.Address) (Info, error)

	//
	//// Delete and DeleteByAddress remove keys.
	//Delete(uid string) error
	//DeleteByAddress(address types.Address) error

	// NewMnemonic generates a new mnemonic, derives a hierarchical deterministic
	// key from that, and persists it to storage. Returns the generated mnemonic and the key
	// Info. It returns an error if it fails to generate a key for the given algo type, or if
	// another key is already stored under the same name.
	NewMnemonic(uid string, language Language, algo SigningAlgo) (Info, string, error)

	// NewAccount converts a mnemonic to a private key and BIP-39 HD Path  and persists it.
	NewAccount(uid, mnemonic, bip39Passwd, hdPath string, algo SigningAlgo) (Info, error)

	//// SaveLedgerKey retrieves a public key reference from a Ledger device and persists it.
	//SaveLedgerKey(uid string, algo SigningAlgo, hrp string, account, index uint32) (Info, error)
	//
	//// SavePubKey stores a public key and returns the persisted Info structure.
	//SavePubKey(uid string, pubkey crypto.PubKey, algo SigningAlgo) (Info, error)
	//
	//// SaveMultisig stores, stores, and returns a new multsig (offline) key reference
	//SaveMultisig(uid string, pubkey crypto.PubKey) (Info, error)
	//
	//// SupportedAlgos returns a list of signing algorithms supported by the keybase
	//SupportedAlgos() []SigningAlgo
	//
	//// SupportedAlgosLedger returns a list of signing algorithms supported by the keybase's ledger integration
	//SupportedAlgosLedger() []SigningAlgo
}

// Signer is implemented by key stores that want to provide signing capabilities.
type Signer interface {
	// Sign and SignByAddress sign byte messages with a user key.
	Sign(uid string, msg []byte) ([]byte, tmcrypto.PubKey, error)
	SignByAddress(address types.Address, msg []byte) ([]byte, tmcrypto.PubKey, error)
}

// Importer is implemented by key stores that support import of public and private keys.
type Importer interface {
	ImportPrivKey(uid, armor, passphrase string) error
	ImportPubKey(uid string, armor string) error
}

// Exporter is implemented by key stores that support export of public and private keys.
type Exporter interface {
	// Export public key
	ExportPubKeyArmor(uid string) (string, error)
	ExportPubKeyArmorByAddress(address types.Address) (string, error)
	// ExportPrivKey returns a private key in ASCII armored format.
	// It returns an error if the key does not exist or a wrong encryption passphrase is supplied.
	ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error)
	ExportPrivKeyArmorByAddress(address types.Address, encryptPassphrase string) (armor string, err error)
}

// NewKeyring creates a new instance of a keyring. Keybase
// options can be applied when generating this new Keybase.
// Available backends are "os", "file", "kwallet", "pass", "test".
func NewAltKeyring(
	appName, backend, rootDir string, userInput io.Reader, opts ...KeybaseOption,
) (Keyring, error) {

	var db keyring.Keyring
	var err error

	switch backend {
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

	return altKeyring{db: db}, nil
}

type altKeyring struct {
	db keyring.Keyring
}

func (a altKeyring) KeyByAddress(address types.Address) (Info, error) {
	ik, err := a.db.Get(string(addrKey(address)))
	if err != nil {
		return nil, err
	}

	if len(ik.Data) == 0 {
		return nil, fmt.Errorf("key with address %s not found", address)
	}

	bs, err := a.db.Get(string(ik.Data))
	if err != nil {
		return nil, err
	}

	return unmarshalInfo(bs.Data)
}

func (a altKeyring) List() ([]Info, error) {
	var res []Info
	keys, err := a.db.Keys()
	if err != nil {
		return nil, err
	}

	sort.Strings(keys)

	for _, key := range keys {
		if strings.HasSuffix(key, infoSuffix) {
			rawInfo, err := a.db.Get(key)
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

func (a altKeyring) Key(uid string) (Info, error) {
	key := infoKey(uid)

	bs, err := a.db.Get(string(key))
	if err != nil {
		return nil, err
	}

	if len(bs.Data) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, uid)
	}

	return unmarshalInfo(bs.Data)
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
