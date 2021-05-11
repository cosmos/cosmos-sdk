package keyring

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	

)

var (
	_ Info = &KeyringEntry{}
)

// Info is the publicly exposed information about a keypair
type Info interface {
	// Human-readable type for key listing
	GetName() string
	// Public key
	GetPubKey() (cryptotypes.PubKey, error)
	// Address
	GetAddress() (types.AccAddress, error)
	// Bip44 Path
	GetPath() (*BIP44Params, error)
}

func NewKeyringEntry(name string, pubKey *codectypes.Any, item isKeyringEntry_Item) *KeyringEntry {
	return &KeyringEntry{name, pubKey, item}
}

func (ke KeyringEntry) GetName() string {
	return ke.Name
}

func (ke KeyringEntry) GetPubKey() (cryptotypes.PubKey, error) {
	pk, ok := ke.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, fmt.Errorf("Unable to cast Pubkey to cryptotypes.PubKey")
	}
	return pk,nil
}

// GetType implements Info interface
func (ke KeyringEntry) GetAddress() (types.AccAddress, error) {
	pk, err := ke.GetPubKey()
	if err != nil {
		return nil, err
	}
	return pk.Address().Bytes(), nil
}

func (ke KeyringEntry) GetPath() (*BIP44Params, error) {
	l := ke.GetLedger()
	switch {
	case l != nil:
		tmp := l.Path
		return tmp, nil
	default:
		return nil, fmt.Errorf("BIP44 Paths are not available for this type")
	}
}

// encoding info
// we remove tis function aso we can pass cdc.Marrshal install ,we put cdc on keystore
/*
func protoMarshalInfo(i Info) ([]byte, error) {
	ke, ok := i.(*KeyringEntry)
	if !ok {
		return nil, fmt.Errorf("Unable to cast Info to *KeyringEntry")
	}

	bz, err := proto.Marshal(ke)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Unable to marshal KeyringEntry to bytes")
	}

	return bz, nil
}
*/

// decoding info
// we remove tis function aso we can pass cdc.Marrshal install ,we put cdc on keystore
/*
func protoUnmarshalInfo(bz []byte, cdc codec.Codec) (Info, error) {

	var ke KeyringEntry // will not work cause we use any, use InterfaceRegistry
	// dont forget to merge master to my branch, UnmarshalBinaryBare has been renamed
	// cdcc.Marshaler.UnmarshalBinaryBare()  // like proto.UnMarshal but works with Any
	if err := cdc.UnmarshalInterface(bz, &ke); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal bytes to Info")
	}

	return ke, nil
}
*/
