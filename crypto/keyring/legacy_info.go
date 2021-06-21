package keyring

import (
	"errors"
	"fmt"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/ledger"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bip39 "github.com/cosmos/go-bip39"
)

// Info is the publicly exposed information about a keypair
type LegacyInfo interface {
	// Human-readable type for key listing
	GetType() KeyType
	// Name of the key
	GetName() string
	// Public key
	GetPubKey() cryptotypes.PubKey
	// Address
	GetAddress() sdk.AccAddress
	// Bip44 Path
	GetPath() (*hd.BIP44Params, error)
	// Algo
	GetAlgo() hd.PubKeyType
}

type LegacyInfoWriter interface {
	// saves legacyInfo of specific accountType to keyring
	NewLegacyMnemonic(uid string, language Language, hdPath, bip39Passphrase string, algo SignatureAlgo, accountType string) (LegacyInfo, string, error)
	NewLegacyAccount(uid, mnemonic, bip39Passphrase, hdPath string, algo SignatureAlgo, accountType string) (LegacyInfo, error)
	// SaveLedgerKey retrieves a public key reference from a Ledger device and persists it.
	SaveLegacyLedgerKey(uid string, algo SignatureAlgo, hrp string, coinType, account, index uint32) (LegacyInfo, error)
	// for testing purposes
	SaveLegacyOfflineKey(uid string, pubkey cryptotypes.PubKey,  algo hd.PubKeyType) (LegacyInfo, error)
	SaveLegacyMultisig(uid string, pubkey cryptotypes.PubKey) (LegacyInfo, error)
}

var (
	_ LegacyInfo = &legacyLocalInfo{}
	_ LegacyInfo = &legacyLedgerInfo{}
	_ LegacyInfo = &legacyOfflineInfo{}
	_ LegacyInfo = &legacyMultiInfo{}
)

// legacyLocalInfo is the public information about a locally stored key
// Note: Algo must be last field in struct for backwards amino compatibility
type legacyLocalInfo struct {
	Name         string             `json:"name"`
	PubKey       cryptotypes.PubKey `json:"pubkey"`
	PrivKeyArmor string             `json:"privkey.armor"`
	Algo         hd.PubKeyType      `json:"algo"`
}

func NewLegacyLocalInfo(name string, pub cryptotypes.PubKey, privArmor string, algo hd.PubKeyType) LegacyInfo {
	return &legacyLocalInfo{
		Name:         name,
		PubKey:       pub,
		PrivKeyArmor: privArmor,
		Algo:         algo,
	}
}

// GetType implements Info interface
func (i legacyLocalInfo) GetType() KeyType {
	return TypeLocal
}

// GetType implements Info interface
func (i legacyLocalInfo) GetName() string {
	return i.Name
}

// GetType implements Info interface
func (i legacyLocalInfo) GetPubKey() cryptotypes.PubKey {
	return i.PubKey
}

// GetType implements Info interface
func (i legacyLocalInfo) GetAddress() sdk.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPrivKeyArmor
func (i legacyLocalInfo) GetPrivKeyArmor() string {
	return i.PrivKeyArmor
}

// GetType implements Info interface
func (i legacyLocalInfo) GetAlgo() hd.PubKeyType {
	return i.Algo
}

// GetType implements Info interface
func (i legacyLocalInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// legacyLedgerInfo is the public information about a Ledger key
// Note: Algo must be last field in struct for backwards amino compatibility
type legacyLedgerInfo struct {
	Name   string             `json:"name"`
	PubKey cryptotypes.PubKey `json:"pubkey"`
	Path   hd.BIP44Params     `json:"path"`
	Algo   hd.PubKeyType      `json:"algo"`
}

func NewLegacyLedgerInfo(name string, pub cryptotypes.PubKey, path hd.BIP44Params, algo hd.PubKeyType) LegacyInfo {
	return &legacyLedgerInfo{
		Name:   name,
		PubKey: pub,
		Path:   path,
		Algo:   algo,
	}
}

// GetType implements Info interface
func (i legacyLedgerInfo) GetType() KeyType {
	return TypeLedger
}

// GetName implements Info interface
func (i legacyLedgerInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i legacyLedgerInfo) GetPubKey() cryptotypes.PubKey {
	return i.PubKey
}

// GetAddress implements Info interface
func (i legacyLedgerInfo) GetAddress() sdk.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPath implements Info interface
func (i legacyLedgerInfo) GetAlgo() hd.PubKeyType {
	return i.Algo
}

// GetPath implements Info interface
func (i legacyLedgerInfo) GetPath() (*hd.BIP44Params, error) {
	tmp := i.Path
	return &tmp, nil
}

// legacyOfflineInfo is the public information about an offline key
// Note: Algo must be last field in struct for backwards amino compatibility
type legacyOfflineInfo struct {
	Name   string             `json:"name"`
	PubKey cryptotypes.PubKey `json:"pubkey"`
	Algo   hd.PubKeyType      `json:"algo"`
}

func newLegacyOfflineInfo(name string, pub cryptotypes.PubKey, algo hd.PubKeyType) LegacyInfo {
	return &legacyOfflineInfo{
		Name:   name,
		PubKey: pub,
		Algo:   algo,
	}
}

// GetType implements Info interface
func (i legacyOfflineInfo) GetType() KeyType {
	return TypeOffline
}

// GetName implements Info interface
func (i legacyOfflineInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i legacyOfflineInfo) GetPubKey() cryptotypes.PubKey {
	return i.PubKey
}

// GetAlgo returns the signing algorithm for the key
func (i legacyOfflineInfo) GetAlgo() hd.PubKeyType {
	return i.Algo
}

// GetAddress implements Info interface
func (i legacyOfflineInfo) GetAddress() sdk.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPath implements Info interface
func (i legacyOfflineInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// Deprecated: this structure is not used anymore and it's here only to allow
// decoding old multiInfo records from keyring.
// The problem with legacy.Cdc.UnmarshalLengthPrefixed - the legacy codec doesn't
// tolerate extensibility.
type multisigPubKeyInfo struct {
	PubKey cryptotypes.PubKey `json:"pubkey"`
	Weight uint               `json:"weight"`
}

// multiInfo is the public information about a multisig key
type legacyMultiInfo struct {
	Name      string               `json:"name"`
	PubKey    cryptotypes.PubKey   `json:"pubkey"`
	Threshold uint                 `json:"threshold"`
	PubKeys   []multisigPubKeyInfo `json:"pubkeys"`
}

// NewMultiInfo creates a new multiInfo instance
func newLegacyMultiInfo(name string, pub cryptotypes.PubKey) (LegacyInfo, error) {
	if _, ok := pub.(*multisig.LegacyAminoPubKey); !ok {
		return nil, fmt.Errorf("MultiInfo supports only multisig.LegacyAminoPubKey, got  %T", pub)
	}
	return &legacyMultiInfo{
		Name:   name,
		PubKey: pub,
	}, nil
}

// GetType implements Info interface
func (i legacyMultiInfo) GetType() KeyType {
	return TypeMulti
}

// GetName implements Info interface
func (i legacyMultiInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i legacyMultiInfo) GetPubKey() cryptotypes.PubKey {
	return i.PubKey
}

// GetAddress implements Info interface
func (i legacyMultiInfo) GetAddress() sdk.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPath implements Info interface
func (i legacyMultiInfo) GetAlgo() hd.PubKeyType {
	return hd.MultiType
}

// GetPath implements Info interface
func (i legacyMultiInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (i legacyMultiInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	multiPK := i.PubKey.(*multisig.LegacyAminoPubKey)

	return codectypes.UnpackInterfaces(multiPK, unpacker)
}

// encoding info
func marshalInfo(i LegacyInfo) []byte {
	return legacy.Cdc.MustMarshalLengthPrefixed(i)
}

// decoding info
func unmarshalInfo(bz []byte) (info LegacyInfo, err error) {
	err = legacy.Cdc.UnmarshalLengthPrefixed(bz, &info)
	if err != nil {
		return nil, err
	}

	// After unmarshalling into &info, if we notice that the info is a
	// multiInfo, then we unmarshal again, explicitly in a multiInfo this time.
	// Since multiInfo implements UnpackInterfacesMessage, this will correctly
	// unpack the underlying anys inside the multiInfo.
	//
	// This is a workaround, as go cannot check that an interface (Info)
	// implements another interface (UnpackInterfacesMessage).
	_, ok := info.(legacyMultiInfo)
	if ok {
		var multi legacyMultiInfo
		err = legacy.Cdc.UnmarshalLengthPrefixed(bz, &multi)

		return multi, err
	}

	return
}

func (ks keystore) NewLegacyMnemonic(uid string, language Language, hdPath, bip39Passphrase string, algo SignatureAlgo, accountType string) (LegacyInfo, string, error) {
	if language != English {
		return nil, "", ErrUnsupportedLanguage
	}

	if !ks.isSupportedSigningAlgo(algo) {
		return nil, "", ErrUnsupportedSigningAlgo
	}

	// Default number of words (24): This generates a mnemonic directly from the
	// number of words by reading system entropy.
	entropy, err := bip39.NewEntropy(DefaultEntropySize)
	if err != nil {
		return nil, "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, "", err
	}

	if bip39Passphrase == "" {
		bip39Passphrase = DefaultBIP39Passphrase
	}

	info, err := ks.NewLegacyAccount(uid, mnemonic, bip39Passphrase, hdPath, algo, accountType)
	if err != nil {
		return nil, "", err
	}

	return info, mnemonic, nil
}

func (ks keystore) NewLegacyAccount(name string, mnemonic string, bip39Passphrase string, hdPath string, algo SignatureAlgo, accountType string) (LegacyInfo, error) {
	if !ks.isSupportedSigningAlgo(algo) {
		return nil, ErrUnsupportedSigningAlgo
	}

	// create master key and derive first key for keyring
	derivedPriv, err := algo.Derive()(mnemonic, bip39Passphrase, hdPath)
	if err != nil {
		return nil, err
	}

	privKey := algo.Generate()(derivedPriv)

	// check if the a key already exists with the same address and return an error
	// if found
	address := sdk.AccAddress(privKey.PubKey().Address())
	if _, err := ks.KeyByAddress(address); err == nil {
		return nil, fmt.Errorf("account with address %s already exists in keyring, delete the key first if you want to recreate it", address)
	}

	var info LegacyInfo

	switch accountType{
	case "local":
		// TODO consider to move logic above line 335 here
		info, err = ks.writeLegacyLocalKey(name, privKey, algo.Name(), accountType)
	case "ledger":
		hrp := "cosmos"
		coinType, account, index := uint32(118), uint32(0), uint32(0) 
		info, err  = ks.SaveLegacyLedgerKey(name, algo, hrp, coinType, account, index)
	case "offline":
		info, err = ks.SaveLegacyOfflineKey(name, privKey.PubKey(), algo.Name())
	case "multi": 
		multi := multisig.NewLegacyAminoPubKey(
			1, []cryptotypes.PubKey{
				privKey.PubKey(),
			},
		)
		info, err = ks.SaveLegacyMultisig(name,  multi)
	}

	return info, err
}

func (ks keystore) writeLegacyLocalKey(name string, priv cryptotypes.PrivKey, algo hd.PubKeyType, accountType string) (LegacyInfo, error) {
	// encrypt private key using keyring
	pub := priv.PubKey()
	info := NewLegacyLocalInfo(name, pub, string(legacy.Cdc.MustMarshal(priv)), algo)
	if err := ks.writeLegacyInfo(info); err != nil {
		return nil, err
	}

	return info, nil
}

func (ks keystore) writeLegacyInfo(info LegacyInfo) error {
	addr := info.GetAddress()
	exists, err := ks.existsInDb(addr, info.GetName())
	if err != nil {
		return err
	}
	if exists {
		return errors.New("public key already exist in keybase")
	}

	key := infoKeyBz(info.GetName())
	serializedInfo := marshalInfo(info)

	err = ks.db.Set(keyring.Item{
		Key:  string(key),
		Data: serializedInfo,
	})
	if err != nil {
		return err
	}

	err = ks.db.Set(keyring.Item{
		Key:  addrHexKeyAsString(addr),
		Data: key,
	})
	if err != nil {
		return err
	}

	return nil
}

// func (ks keystore) ExportPrivateKeyFromLegacyInfo(uid string) (types.PrivKey, error) {
func exportPrivateKeyFromLegacyInfo(info LegacyInfo) (cryptotypes.PrivKey, error) {

	switch linfo := info.(type) {
	case legacyLocalInfo:
		if linfo.PrivKeyArmor == "" {
			return nil, fmt.Errorf("private key not available")
		}

		priv, err := legacy.PrivKeyFromBytes([]byte(linfo.PrivKeyArmor))
		if err != nil {
			return nil, err
		}

		return priv, nil
	//case legacyLedgerInfo, legacyOfflineInfo, legacyMultiInfo:
	default:
		return nil, errors.New("only works on local private keys")
	}
}

func (ks keystore) SaveLegacyLedgerKey(uid string, algo SignatureAlgo, hrp string, coinType, account, index uint32) (LegacyInfo, error) {

	if !ks.options.SupportedAlgosLedger.Contains(algo) {
		return nil, ErrUnsupportedSigningAlgo
	}

	hdPath := hd.NewFundraiserParams(account, coinType, index)

	priv, _, err := ledger.NewPrivKeySecp256k1(*hdPath, hrp)
	if err != nil {
		return nil, err
	}

	return ks.writeLegacyLedgerKey(uid, priv.PubKey(), hdPath, algo.Name())
}

func (ks keystore) writeLegacyLedgerKey(name string, pub cryptotypes.PubKey, path *hd.BIP44Params, algo hd.PubKeyType) (LegacyInfo, error) {
	info := NewLegacyLedgerInfo(name, pub, *path, algo)
	if err := ks.writeLegacyInfo(info); err != nil {
		return nil, err
	}

	return info, nil
}

func (ks keystore) SaveLegacyOfflineKey(uid string, pubkey cryptotypes.PubKey, algo hd.PubKeyType) (LegacyInfo, error) {
	return ks.writeLegacyOfflineKey(uid, pubkey, algo)
}

func (ks keystore) writeLegacyOfflineKey(name string, pub cryptotypes.PubKey, algo hd.PubKeyType) (LegacyInfo, error) {
	info := newLegacyOfflineInfo(name, pub, algo)
	err := ks.writeLegacyInfo(info)
	if err != nil {
		return nil, err
	}

	return info, nil
}


func (ks keystore) SaveLegacyMultisig(uid string, pubkey cryptotypes.PubKey) (LegacyInfo, error) {
	return ks.writeLegacyMultisigKey(uid, pubkey)
}

func (ks keystore) writeLegacyMultisigKey(name string, pub cryptotypes.PubKey) (LegacyInfo, error) {
	info, err := newLegacyMultiInfo(name, pub)
	if err != nil {
		return nil, err
	}
	if err = ks.writeLegacyInfo(info); err != nil {
		return nil, err
	}

	return info, nil
}