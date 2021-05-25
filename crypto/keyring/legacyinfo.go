package keyring

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
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
	GetAddress() types.AccAddress
	// Bip44 Path
	GetPath() (*hd.BIP44Params, error)
	// Algo
	GetAlgo() hd.PubKeyType
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

func newLegacyLocalInfo(name string, pub cryptotypes.PubKey, privArmor string, algo hd.PubKeyType) LegacyInfo {
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
func (i legacyLocalInfo) GetAddress() types.AccAddress {
	return i.PubKey.Address().Bytes()
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

func newLegacyLedgerInfo(name string, pub cryptotypes.PubKey, path hd.BIP44Params, algo hd.PubKeyType) LegacyInfo {
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
func (i legacyLedgerInfo) GetAddress() types.AccAddress {
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

func newlegacyOfflineInfo(name string, pub cryptotypes.PubKey, algo hd.PubKeyType) LegacyInfo {
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
func (i legacyOfflineInfo) GetAddress() types.AccAddress {
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
func NewMultiInfo(name string, pub cryptotypes.PubKey) (LegacyInfo, error) {
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
func (i legacyMultiInfo) GetAddress() types.AccAddress {
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