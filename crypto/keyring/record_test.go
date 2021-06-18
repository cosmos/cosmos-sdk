package keyring_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func TestEmptyRecordMarshaling(t *testing.T) {
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


func TestLocalRecordMarshaling(t *testing.T) {
	require := require.New(t)

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()
	
	var privKey cryptotypes.PrivKey
	privKey = priv
	
	localRecord, err := keyring.NewLocalRecord(cdc, privKey)
	require.NoError(err)
	localRecordItem := keyring.NewLocalRecordItem(localRecord)

	r, err := keyring.NewRecord("testrecord", pub, localRecordItem)
	require.NoError(err)

	bz, err := cdc.Marshal(r)
	require.NoError(err)

	var r2 keyring.Record
	require.NoError(cdc.Unmarshal(bz, &r2))
	require.Equal(r.Name, r2.Name)
	// not sure if this will work -- we can remove this line, the later check is better.
	require.True(r.PubKey.Equal(r2.PubKey))

	pub2, err := r2.GetPubKey()
	require.NoError(err)
	require.True(pub.Equals(pub2))

	localRecord2 := r2.GetLocal()
	bzPriv, err := cdc.Marshal(priv)
	require.NoError(err)
	require.Equal(localRecord2.PrivKeyArmor, string(bzPriv))
	require.Equal(localRecord2.PrivKeyType, privKey.Type())
}




/* TODO implement tests
TestNewRecordGetItem
input: name, anyPub, a)Empty b)Local 3)ledger
input: any
call NewRecord()
test GetLocal, GetLedger, GetEmpty


func newLocalRecord(cdc codec.Codec, privKey cryptotypes.PrivKey) (*Record_Local, error) {
TestNewLocalRecord 
input privKey is valid and invalid

test extractPrivKeyFrom Local




*/
