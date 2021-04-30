package keyring

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/ptypes"
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

func (ke KeyringEntry) GetName() string {
	return ke.Name
}

func (ke KeyringEntry) GetPubKey() (pk cryptotypes.PubKey, err error) {
	return ke.unmarshalAnytoPubKey()
}

// GetType implements Info interface
func (ke KeyringEntry) GetAddress() (types.AccAddress, error) {
	pk, err := ke.unmarshalAnytoPubKey()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal any to Pubkey")
	}
	return pk.Address().Bytes(),nil
}

func (ke *KeyringEntry) GetPath() (*BIP44Params, error) {
	switch {
	case ke.GetLedger() != nil:
		l := ke.GetLedger() 
		tmp := l.Path
		return tmp, nil
	case ke.GetOffline() != nil, ke.GetLocal() != nil, ke.GetMulti() != nil:               
		return nil, fmt.Errorf("BIP44 Paths are not available for this type")
	}
	/* 
		should I handle this case? or just the better approach is:
		if ke.GetLedger() != nil {
			l := ke.GetLedger() 
			tmp := l.Path
			return tmp, nil
		}
		return nil, fmt.Errorf("BIP44 Paths are not available for this type")

	*/

	return nil, fmt.Errorf("some error")  
}


func (ke KeyringEntry) unmarshalAnytoPubKey() (pk cryptotypes.PubKey, err error) {
	if err := ptypes.UnmarshalAny(ke.PubKey, &pk); err != nil {
		return nil, err
	}
	return
}

// encoding info
func protoMarshalInfo(i Info) ([]byte, error) {
	ke, ok := i.(*KeyringEntry) // address error
	bz, _ := proto.Marshal(ke) // address error
	return bz

}

// decoding info
func protoUnmarshalInfo(bz []byte, cdc codec.Codec) (Info, error) {
	// first merge master to my branch
	var ke KeyringEntry // will not work cause we use any, use InterfaceRegistry
	// cdcc.Marshaler.UnmarshalBinaryBare()  // like proto.UnMarshal but works with Any
	if err := cdc.Unmarshal(bz, &ke); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal bytes to Info")
	}

	return ke, nil
}
