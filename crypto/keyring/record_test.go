package keyring_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
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
// TODO fix that
func TestExtractPrivKeyFromItem(t *testing.T){
	
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	tt := []struct {
		name string
		privKey cryptotypes.PrivKey
		errExp  error
	}{
		{
			name: "local record",
			privKey: ed25519.GenPrivKey(),
			errExp : nil,
		},
		{
			name: "ledger record",
			privKey: secp256k1.GenPrivKey(),
			errExp : keyring.ErrPrivKeyExtr,
		},
		{
			name: "empty record",
			privKey: ed25519.GenPrivKey(),
			errExp : keyring.ErrPrivKeyExtr,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name,  func(t *testing.T) {
			require := require.New(t)
			k := new(keyring.Record)
		
			switch tc.name {
			case "local record":
				localRecord, err := keyring.NewLocalRecord(cdc, tc.privKey)
				require.NoError(err)
				localRecordItem := keyring.NewLocalRecordItem(localRecord)
				k, err = keyring.NewRecord(tc.name, tc.privKey.PubKey(), localRecordItem)
				require.NoError(err)
			case "ledger record":
				var err error
				path := hd.NewFundraiserParams(4, types.CoinType, 22)
				ledgerRecord := keyring.NewLedgerRecord(path)
				ledgerRecordItem := keyring.NewLedgerRecordItem(ledgerRecord)
				k, err = keyring.NewRecord(tc.name, tc.privKey.PubKey(), ledgerRecordItem)
				require.NoError(err)
			case "empty record":
				var err error
				emptyRecord := keyring.NewEmptyRecord()
				emptyRecordItem := keyring.NewEmptyRecordItem(emptyRecord)
				k, err = keyring.NewRecord(tc.name, tc.privKey.PubKey(), emptyRecordItem)
				require.NoError(err)
			}

			priv, err := keyring.ExtractPrivKeyFromItem(cdc, k)
			require.Equal(tc.errExp, err)
			require.Equal(priv, tc.privKey)

		})
	}
}