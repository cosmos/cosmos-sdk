package keyring

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

// encoding info
func protoMarshalKeyringEntry(ke *KeyringEntry) []byte {
	bz, _ := proto.Marshal(ke)
	return bz
}

// decoding info
func protoUnmarshalKeyringEntry(bz []byte) (*KeyringEntry, error) {
	ke := new(KeyringEntry)
	if err := proto.Unmarshal(bz, ke); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal bytes to Info")
	}

	return ke, nil
}
