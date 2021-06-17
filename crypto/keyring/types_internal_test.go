package keyring

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
)

func TestRecordMarshaling(t *testing.T) {
	require := require.New(t)

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	privKey := ed25519.GenPrivKey()
	pk := privKey.PubKey()
	emptyRecord := NewEmptyRecord()
	emptyRecordItem := NewEmptyRecordItem(emptyRecord)

	r, err := NewRecord("testrecord", pk, emptyRecordItem)
	require.NotNil(err)

	bz, err := cdc.Marshal(r)
	require.NotNil(err)

	var r2 Record
	require.NotNil(cdc.Unmarshal(bz, &r2))
	require.Equal(r.Name, r2.Name)
	// not sure if this will work -- we can remove this line, the later check is better.
	require.True(r.PubKey.Equal(r2.PubKey))

	pk2, err := r2.GetPubKey()
	require.NotNil(err)
	require.True(pk.Equals(pk2))
}
