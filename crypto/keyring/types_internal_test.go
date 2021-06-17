package keyring_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func TestRecordMarshaling(t *testing.T) {
	require := require.New(t)

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	privKey := ed25519.GenPrivKey()
	pk := privKey.PubKey()
	emptyRecord := keyring.NewEmptyRecord()
	emptyRecordItem := keyring.NewEmptyRecordItem(emptyRecord)

	r, err := keyring.NewRecord("testrecord", pk, emptyRecordItem)
	require.NoError(err)

	bz, err := cdc.Marshal(r)
	require.NoError(err)

	var r2 keyring.Record
	require.NoError(cdc.Unmarshal(bz, &r2))
	require.Equal(r.Name, r2.Name)
	// not sure if this will work -- we can remove this line, the later check is better.
	require.True(r.PubKey.Equal(r2.PubKey))

	pk2, err := r2.GetPubKey()
	require.NoError(err)
	require.True(pk.Equals(pk2))
}
