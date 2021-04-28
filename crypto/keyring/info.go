package keyring

import (
	"fmt"

	crypto_ed25519 "crypto/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

// Info is the publicly exposed information about a keypair

// TODO fix GetPath and GetAlgo in whole codebase
type Info interface {
	// Human-readable type for key listing
	GetType() KeyType
	// Name of the key
	GetName() string
	// Public key
	GetPubKey() crypto_ed25519.PublicKey
	// Address
	GetAddress() types.AccAddress
	// Bip44 Path
	GetPath() (*BIP44Params, error)
	// Algo
	GetAlgo() PubKeyType
}

var (
	_ Info = &LocalInfo{}
	_ Info = &LedgerInfo{}
	_ Info = &OfflineInfo{}
	_ Info = &MultiInfo{}
)

// TODO fix errors, GetAddress and GetAlgo

//LocalInfo

func (i LocalInfo) String() KeyType {
	return fmt.Sprintf("LocalInfo{%s}", i.Name)
}

// GetType implements Info interface
func (i LocalInfo) GetType() KeyType {
	// review Type
	return TypeLocal
}

// GetType implements Info interface
func (i LocalInfo) GetName() string {
	return i.Name
}

// GetType implements Info interface
func (i LocalInfo) GetPubKey() crypto_ed25519.PublicKey {
	return i.Key
}

// GetType implements Info interface
func (i LocalInfo) GetAddress() types.AccAddress {
	return i.Key.Address().Bytes()
}

// GetType implements Info interface
func (i LocalInfo) GetAlgo() PubKeyType  {
	return i.Algo
}

// GetType implements Info interface
func (i LocalInfo) GetPath() (*BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// LEDGERINFO

func (i LedgerInfo) GetType() KeyType {
	return TypeLedger
}

// GetName implements Info interface
func (i LedgerInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i LedgerInfo) GetPubKey() crypto_ed25519.PublicKey {
	return i.Key
}

// GetAddress implements Info interface
func (i LedgerInfo) GetAddress() types.AccAddress {
	return i.Key.Address().Bytes()
}

// GetPath implements Info interface
func (i LedgerInfo) GetAlgo() PubKeyType {
	return i.Algo
}

// GetPath implements Info interface
func (i LedgerInfo) GetPath() (*BIP44Params, error) {
	tmp := i.Path
	return &tmp, nil
}


// OFFLINEINFO

func (i OfflineInfo) GetType() KeyType {
	return TypeOffline
}

// GetName implements Info interface
func (i OfflineInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i OfflineInfo) GetPubKey() crypto_ed25519.PublicKey {
	return i.Key
}

// GetAlgo returns the signing algorithm for the key
func (i OfflineInfo) GetAlgo() PubKeyType {
	return i.Algo
}

// GetAddress implements Info interface
func (i OfflineInfo) GetAddress() types.AccAddress {
	return i.Key.Address().Bytes()
}

// GetPath implements Info interface
func (i OfflineInfo) GetPath() (*BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// MULTIINFO

// GetType implements Info interface
func (i MultiInfo) GetType() KeyType {
	return TypeMulti
}

// GetName implements Info interface
func (i MultiInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i MultiInfo) GetPubKey() crypto_ed25519.PublicKey {
	return i.Key
}

// GetAddress implements Info interface
func (i MultiInfo) GetAddress() types.AccAddress {
	return i.Key.Address().Bytes()
}

// GetPath implements Info interface
func (i MultiInfo) GetAlgo() PubKeyType {
	return PubKeyType_MultiType
}

// GetPath implements Info interface
func (i MultiInfo) GetPath() (*BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// encoding info
func protoMarshalInfo(i Info) []byte {
	bz, _ := proto.Marshal(i.(proto.Message))
	return bz
}

// decoding info
func protoUnmarshalInfo(bz []byte) (info Info, err error) {
	if err := proto.Unmarshal(bz, info.(proto.Message)); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal bytes to Info")
	}

	return
}
