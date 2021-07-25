package keyring_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
)

func TestOfflineRecordMarshaling(t *testing.T) {
	require := require.New(t)

	privKey := ed25519.GenPrivKey()
	pk := privKey.PubKey()
	emptyRecord := keyring.NewOfflineRecord()
	emptyRecordItem := keyring.NewOfflineRecordItem(emptyRecord)

	r, err := keyring.NewRecord("testrecord", pk, emptyRecordItem)
	require.NoError(err)

	cdc := simapp.MakeTestEncodingConfig().Codec
	bz, err := cdc.Marshal(r)
	require.NoError(err)

	var r2 keyring.Record
	require.NoError(cdc.Unmarshal(bz, &r2))
	require.Equal(r.Name, r2.Name)
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

	localRecord2 := r2.GetLocal()
	require.NotNil(localRecord2)
	anyPrivKey, err := codectypes.NewAnyWithValue(privKey)
	require.NoError(err)
	require.Equal(localRecord2.PrivKey, anyPrivKey)
	require.Equal(localRecord2.PrivKeyType, privKey.Type())
}

func TestLedgerRecordMarshaling(t *testing.T) {

	const n1 = "cosmos"
	require := require.New(t)
	dir := t.TempDir()
	mockIn := strings.NewReader("")

	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()

	encCfg := simapp.MakeTestEncodingConfig()
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	path := hd.NewFundraiserParams(4, 12345, 57)
	ledgerRecord := keyring.NewLedgerRecord(path)
	require.NoError(err)
	ledgerRecordItem := keyring.NewLedgerRecordItem(ledgerRecord)

	r, err := keyring.NewRecord("testrecord", pub, ledgerRecordItem)
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

	ledgerRecord2 := r2.GetLedger()
	require.NotNil(ledgerRecord2)
	require.Nil(r2.GetLocal())

	require.Equal(ledgerRecord2.Path.String(), path.String())
}

func TestExtractPrivKeyFromLocalRecord(t *testing.T) {
	require := require.New(t)

	priv := secp256k1.GenPrivKey()
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

func TestExtractPrivKeyFromOfflineRecord(t *testing.T) {
	require := require.New(t)

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey()

	offlineRecord := keyring.NewOfflineRecord()
	emptyRecordItem := keyring.NewOfflineRecordItem(offlineRecord)

	k, err := keyring.NewRecord("testrecord", pub, emptyRecordItem)
	require.NoError(err)

	privKey2, err := keyring.ExtractPrivKeyFromRecord(k)
	require.Error(err)
	require.Nil(privKey2)
}
