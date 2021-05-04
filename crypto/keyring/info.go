package keyring

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

var (
	_ Info = &KeyringEntry{}
)

// Info is the publicly exposed information about a keypair
type Info interface {
	// Human-readable type for key listing
	GetName() string
	// Public key
	GetPubKey(cdc codec.AminoCodec) (cryptotypes.PubKey, error)
	// Address
	GetAddress(cdc codec.AminoCodec) (types.AccAddress, error)
	// Bip44 Path
	GetPath() (*BIP44Params, error)
}

func (ke KeyringEntry) GetName() string {
	return ke.Name
}

func (ke KeyringEntry) GetPubKey(cdc codec.AminoCodec) (pk cryptotypes.PubKey, err error) {
	return ke.unmarshalAnytoPubKey(cdc)
}

// GetType implements Info interface
func (ke KeyringEntry) GetAddress(cdc codec.AminoCodec) (types.AccAddress, error) {
	pk, err := ke.unmarshalAnytoPubKey(cdc)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal any to Pubkey")
	}
	return pk.Address().Bytes(),nil
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

func (ke KeyringEntry) unmarshalAnytoPubKey(cdc codec.AminoCodec) (pk cryptotypes.PubKey, err error) {
	if err := cdc.UnmarshalInterface(ke.PubKey, &pk); err != nil{
		return nil, err
	}
	return
}

// encoding info
func protoMarshalInfo(i Info) ([]byte, error) {
	ke, ok := i.(*KeyringEntry)
	if !ok {
		return nil, fmt.Errorf("Unable to cast Info to *KeyringEntry")
	}

	bz, err := proto.Marshal(ke)
	if err != nil {
		return nil,sdkerrors.Wrap(err, "Unable to marshal KeyringEntry to bytes")
	}

	return bz,nil
}

// decoding info
func protoUnmarshalInfo(bz []byte, cdc codec.AminoCodec) (Info, error) {
	
	var ke KeyringEntry // will not work cause we use any, use InterfaceRegistry
	// dont forget to merge master to my branch, UnmarshalBinaryBare has been renamed
	// cdcc.Marshaler.UnmarshalBinaryBare()  // like proto.UnMarshal but works with Any
	if err := cdc.UnmarshalInterface(bz, &ke); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal bytes to Info")
	}

	return ke, nil
}
