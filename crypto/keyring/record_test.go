package keyring_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
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

	const n1 = "cosmos"
	require := require.New(t)
	dir := t.TempDir()
	mockIn := strings.NewReader("")

	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()

	var privKey cryptotypes.PrivKey
	privKey = priv

	encCfg := simapp.MakeTestEncodingConfig()
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	localRecord, err := keyring.NewLocalRecord(privKey)
	require.NoError(err)
	localRecordItem := keyring.NewLocalRecordItem(localRecord)

	r, err := keyring.NewRecord("testrecord", pub, localRecordItem)
	require.NoError(err)

	bz, err := kb.ProtoMarshalRecord(r)
	require.NoError(err)

	r2, err := kb.ProtoUnmarshalRecord(bz)
	require.NoError(err)
	require.Equal(r.Name, r2.Name)
	// not sure if this will work -- we can remove this line, the later check is better.
	require.True(r.PubKey.Equal(r2.PubKey))

	pub2, err := r2.GetPubKey()
	require.NoError(err)
	require.True(pub.Equals(pub2))

	/* TODO fix that
	localRecord2 := r2.GetLocal()
	bzPriv, err := kb.MarshalPrivKey(priv)
	require.NoError(err)
	require.Equal(localRecord2.PrivKeyArmor, string(bzPriv))
	require.Equal(localRecord2.PrivKeyType, privKey.Type())
	*/
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
// add test with Marshalling of Record
func TestExtractPrivKeyFromLocalRecord(t *testing.T) {
	require := require.New(t)

	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()
	privKey := cryptotypes.PrivKey(priv)
	
	// use proto serialize
	localRecord, err := keyring.NewLocalRecord(privKey)
	require.NoError(err)
	localRecordItem := keyring.NewLocalRecordItem(localRecord)

	k, err := keyring.NewRecord("testrecord", pub, localRecordItem)
	require.NoError(err)

	

	privKey2, err := keyring.ExtractPrivKeyFromRecord(k)
	require.NoError(err)
	require.True(privKey2.Equals(privKey))
}

func TestExtractPrivKeyFromEmptyRecord(t *testing.T) {
	require := require.New(t)

	/*
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)
	*/

	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()

	emptyRecord := keyring.NewEmptyRecord()
	emptyRecordItem := keyring.NewEmptyRecordItem(emptyRecord)

	k, err := keyring.NewRecord("testrecord", pub, emptyRecordItem)
	require.NoError(err)

	privKey2, err := keyring.ExtractPrivKeyFromRecord(k)
	require.Error(err)
	require.Nil(privKey2)
}


// TODO fix that
/*
func TestExtractPrivKeyFromRecord(t *testing.T){

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

			_, err := keyring.ExtractPrivKeyFromRecord(cdc, k)
			require.Equal(tc.errExp, err)
	        // TODO find out how to compare 2 private keys
			//require.True(bytes.Equal(priv.Key., tc.privKey.PubKey().String())

		})
	}
}
*/
