package keyring

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/ptypes"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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
	GetPubKey() *codectypes.Any
	// Address
	GetAddress() types.AccAddress
	// Bip44 Path
	GetPath() (*BIP44Params, error)
}

func (ke KeyringEntry) GetName() string {
	return ke.Name
}

func (ke KeyringEntry) GetPubKey() *codectypes.Any {
	return ke.PubKey
}

// GetType implements Info interface
func (ke KeyringEntry) GetAddress() types.AccAddress {
	var pk cryptotypes.PubKey
	ptypes.UnmarshalAny(ke.PubKey, &pk) // handle error or not?
	return pk.Address().Bytes()
}

func (ke *KeyringEntry) GetPath() (*BIP44Params, error) {
	switch {
	case ke.GetLedger() != nil:
		l := ke.GetLedger() 
		tmp := l.Path
		return tmp, nil
	case ke.GetOffline() != nil, ke.GetLocal() != nil, ke.GetMulti() != nil:                :
		return nil, fmt.Errorf("BIP44 Paths are not available for this type")
	}
}


// encoding info
func protoMarshalInfo(i Info) []byte {
	ke, _ := i.(*KeyringEntry)
	bz, _ := proto.Marshal(ke)
	return bz

	/*
	or
	bz, _ := proto.Marshal(i.(proto.Message))
	return bz
	*/
}

// decoding info
func protoUnmarshalInfo(bz []byte) (Info, error) {
	var ke KeyringEntry
	if err := proto.Unmarshal(bz, &ke); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal bytes to Info")
	}

	return ke, nil
}
