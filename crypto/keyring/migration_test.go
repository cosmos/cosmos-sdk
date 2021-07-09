package keyring_test

import (
	"strings"
	"testing"

	design99keyring "github.com/99designs/keyring"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
)

const n1 = "cosmos"

// TODO consider to make table driven testMigrationLegacy tests
// TODO test MigrateAll
func TestMigrateLegacyLocalKey(t *testing.T) {
	//saves legacyLocalInfo to keyring
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	privKey := cryptotypes.PrivKey(priv)
	pub := priv.PubKey()

	// TODO serialize using amino or proto? legacy.Cdc.MustMarshal(priv)
	legacyLocalInfo := keyring.NewLegacyLocalInfo(n1, pub, string(legacy.Cdc.MustMarshal(privKey)), hd.Secp256k1.Name())
	serializedLegacyLocalInfo := keyring.MarshalInfo(legacyLocalInfo)

	item := design99keyring.Item{
		Key:         n1,
		Data:        serializedLegacyLocalInfo,
		Description: "SDK kerying version",
	}

	err = kb.SetItem(item)
	require.NoError(err)

	migrated, err := kb.Migrate(n1)
	require.True(migrated)
	require.NoError(err)
}

// test pass!
// go test -tags='cgo ledger norace' github.com/cosmos-sdk/crypto
func TestMigrateLegacyLedgerKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey()

	account, coinType, index := uint32(118), uint32(0), uint32(0)
	hdPath := hd.NewFundraiserParams(account, coinType, index)
	legacyLedgerInfo := keyring.NewLegacyLedgerInfo(n1, pub, *hdPath, hd.Secp256k1.Name())
	serializedLegacyLedgerInfo := keyring.MarshalInfo(legacyLedgerInfo)

	item := design99keyring.Item{
		Key:         n1,
		Data:        serializedLegacyLedgerInfo,
		Description: "SDK kerying version",
	}

	err = kb.SetItem(item)
	require.NoError(err)

	migrated, err := kb.Migrate(n1)
	require.True(migrated)
	require.NoError(err)
}

func TestMigrateLegacyOfflineKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey()

	legacyOfflineInfo := keyring.NewLegacyOfflineInfo(n1, pub, hd.Secp256k1.Name())
	serializedLegacyOfflineInfo := keyring.MarshalInfo(legacyOfflineInfo)

	item := design99keyring.Item{
		Key:         n1,
		Data:        serializedLegacyOfflineInfo,
		Description: "SDK kerying version",
	}

	err = kb.SetItem(item)
	require.NoError(err)

	migrated, err := kb.Migrate(n1)
	require.True(migrated)
	require.NoError(err)
}

func TestMigrateLegacyMultiKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	multi := multisig.NewLegacyAminoPubKey(
		1, []cryptotypes.PubKey{
			priv.PubKey(),
		},
	)
	legacyMultiInfo, err := keyring.NewLegacyMultiInfo(n1, multi)
	require.NoError(err)
	serializedLegacyMultiInfo := keyring.MarshalInfo(legacyMultiInfo)

	item := design99keyring.Item{
		Key:         n1,
		Data:        serializedLegacyMultiInfo,
		Description: "SDK kerying version",
	}

	err = kb.SetItem(item)
	require.NoError(err)

	migrated, err := kb.Migrate(n1)
	require.True(migrated)
	require.NoError(err)
}

// TODO  do i need to test migration for ledger,offline record items as well?
// TODO make keystore.Cdc field public

func TestMigrateLocalRecord(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := simapp.MakeTestEncodingConfig().Codec

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	privKey := cryptotypes.PrivKey(priv)
	pub := priv.PubKey()

	localRecord, err := keyring.NewLocalRecord(privKey)
	require.NoError(err)
	localRecordItem := keyring.NewLocalRecordItem(localRecord)
	k, err := keyring.NewRecord("test record", pub, localRecordItem)
	serializedRecord, err := kb.ProtoMarshalRecord(k)
	require.NoError(err)

	item := design99keyring.Item{
		Key:         n1,
		Data:        serializedRecord,
		Description: "SDK kerying version",
	}

	err = kb.SetItem(item)
	require.NoError(err)

	migrated, err := kb.Migrate(n1)
	require.False(migrated)
	require.NoError(err)
}

// TODO insert multiple incorrect migration keys and output errors to user
func TestMigrateOneRandomItemError(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	randomBytes := []byte("abckd0s03l")

	errItem := design99keyring.Item{
		Key:         n1,
		Data:        randomBytes,
		Description: "SDK kerying version",
	}

	err = kb.SetItem(errItem)
	require.NoError(err)

	migrated, err := kb.Migrate(n1)
	require.False(migrated)
	require.Error(err)
}

func TestMigrateAllMultiOffline(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	multi := multisig.NewLegacyAminoPubKey(
		1, []cryptotypes.PubKey{
			priv.PubKey(),
		},
	)
	legacyMultiInfo, err := keyring.NewLegacyMultiInfo(n1, multi)
	require.NoError(err)
	serializedLegacyMultiInfo := keyring.MarshalInfo(legacyMultiInfo)

	item := design99keyring.Item{
		Key:         n1,
		Data:        serializedLegacyMultiInfo,
		Description: "SDK kerying version",
	}

	require.NoError(kb.SetItem(item))

	priv = secp256k1.GenPrivKey()
	pub := priv.PubKey()

	legacyOfflineInfo := keyring.NewLegacyOfflineInfo(n1, pub, hd.Secp256k1.Name())
	serializedLegacyOfflineInfo := keyring.MarshalInfo(legacyOfflineInfo)

	item = design99keyring.Item{
		Key:         n1,
		Data:        serializedLegacyOfflineInfo,
		Description: "SDK kerying version",
	}

	require.NoError(kb.SetItem(item))

	migrated, err := kb.MigrateAll()
	require.True(migrated)
	require.NoError(err)

}

func TestMigrateAllNoItem(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	migrated, err := kb.MigrateAll()
	require.False(migrated)
	require.NoError(err)
}

func TestMigrateErrUnknownItemKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Codec)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey()

	legacyOfflineInfo := keyring.NewLegacyOfflineInfo(n1, pub, hd.Secp256k1.Name())
	serializedLegacyOfflineInfo := keyring.MarshalInfo(legacyOfflineInfo)

	item := design99keyring.Item{
		Key:         n1,
		Data:        serializedLegacyOfflineInfo,
		Description: "SDK kerying version",
	}

	err = kb.SetItem(item)
	require.NoError(err)

	incorrectItemKey := n1 + "1"
	migrated, err := kb.Migrate(incorrectItemKey)
	require.False(migrated)
	require.True(strings.Contains(err.Error(), "key not found"))
}
