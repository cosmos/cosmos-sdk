package keyring

import (
	"strings"
	"testing"

	"github.com/99designs/keyring"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/hd"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const n1 = "cosmos"

func TestMigrateLegacyLocalKey(t *testing.T) {
	//saves legacyLocalInfo to keyring
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	privKey := cryptotypes.PrivKey(priv)
	pub := priv.PubKey()

	legacyLocalInfo := NewLegacyLocalInfo(n1, pub, string(legacy.Cdc.MustMarshal(privKey)), hd.Secp256k1.Name())
	serializedLegacyLocalInfo := MarshalInfo(legacyLocalInfo)

	item := keyring.Item{
		Key:         n1,
		Data:        serializedLegacyLocalInfo,
		Description: "SDK kerying version",
	}

	ks, ok := kb.(keystore)
	require.True(ok)
	require.NoError(ks.setItem(item))

	migrated, err := ks.migrate(n1)
	require.True(migrated)
	require.NoError(err)
}

// test pass!
// go test -tags='cgo ledger norace' github.com/cosmos-sdk/crypto

func TestMigrateLegacyLedgerKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey()

	account, coinType, index := uint32(118), uint32(0), uint32(0)
	hdPath := hd.NewFundraiserParams(account, coinType, index)
	legacyLedgerInfo := NewLegacyLedgerInfo(n1, pub, *hdPath, hd.Secp256k1.Name())
	serializedLegacyLedgerInfo := MarshalInfo(legacyLedgerInfo)

	item := keyring.Item{
		Key:         n1,
		Data:        serializedLegacyLedgerInfo,
		Description: "SDK kerying version",
	}

	ks, ok := kb.(keystore)
	require.True(ok)
	require.NoError(ks.setItem(item))

	migrated, err := ks.migrate(n1)
	require.True(migrated)
	require.NoError(err)
}

func TestMigrateLegacyOfflineKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey()

	legacyOfflineInfo := NewLegacyOfflineInfo(n1, pub, hd.Secp256k1.Name())
	serializedLegacyOfflineInfo := MarshalInfo(legacyOfflineInfo)

	item := keyring.Item{
		Key:         n1,
		Data:        serializedLegacyOfflineInfo,
		Description: "SDK kerying version",
	}

	ks, ok := kb.(keystore)
	require.True(ok)
	require.NoError(ks.setItem(item))

	migrated, err := ks.migrate(n1)
	require.True(migrated)
	require.NoError(err)
}

func TestMigrateLegacyMultiKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	multi := multisig.NewLegacyAminoPubKey(
		1, []cryptotypes.PubKey{
			priv.PubKey(),
		},
	)
	legacyMultiInfo, err := NewLegacyMultiInfo(n1, multi)
	require.NoError(err)
	serializedLegacyMultiInfo := MarshalInfo(legacyMultiInfo)

	item := keyring.Item{
		Key:         n1,
		Data:        serializedLegacyMultiInfo,
		Description: "SDK kerying version",
	}

	ks, ok := kb.(keystore)
	require.True(ok)
	require.NoError(ks.setItem(item))

	migrated, err := ks.migrate(n1)
	require.True(migrated)
	require.NoError(err)
}

func TestMigrateLocalRecord(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	privKey := cryptotypes.PrivKey(priv)
	pub := priv.PubKey()

	localRecord, err := NewLocalRecord(privKey)
	require.NoError(err)
	localRecordItem := NewLocalRecordItem(localRecord)
	k, err := NewRecord("test record", pub, localRecordItem)

	ks, ok := kb.(keystore)
	require.True(ok)

	serializedRecord, err := ks.protoMarshalRecord(k)
	require.NoError(err)

	item := keyring.Item{
		Key:         n1,
		Data:        serializedRecord,
		Description: "SDK kerying version",
	}

	require.NoError(ks.setItem(item))

	migrated, err := ks.migrate(n1)
	require.False(migrated)
	require.NoError(err)
}

func TestMigrateOneRandomItemError(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	randomBytes := []byte("abckd0s03l")

	errItem := keyring.Item{
		Key:         n1,
		Data:        randomBytes,
		Description: "SDK kerying version",
	}

	ks, ok := kb.(keystore)
	require.True(ok)
	require.NoError(ks.setItem(errItem))

	migrated, err := ks.migrate(n1)
	require.False(migrated)
	require.Error(err)
}

func TestMigrateAllMultiOffline(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	multi := multisig.NewLegacyAminoPubKey(
		1, []cryptotypes.PubKey{
			priv.PubKey(),
		},
	)
	legacyMultiInfo, err := NewLegacyMultiInfo(n1, multi)
	require.NoError(err)
	serializedLegacyMultiInfo := MarshalInfo(legacyMultiInfo)

	item := keyring.Item{
		Key:         n1,
		Data:        serializedLegacyMultiInfo,
		Description: "SDK kerying version",
	}

	ks, ok := kb.(keystore)
	require.True(ok)
	require.NoError(ks.setItem(item))

	priv = secp256k1.GenPrivKey()
	pub := priv.PubKey()

	legacyOfflineInfo := NewLegacyOfflineInfo(n1, pub, hd.Secp256k1.Name())
	serializedLegacyOfflineInfo := MarshalInfo(legacyOfflineInfo)

	item = keyring.Item{
		Key:         n1,
		Data:        serializedLegacyOfflineInfo,
		Description: "SDK kerying version",
	}

	require.NoError(ks.setItem(item))

	migrated, err := kb.MigrateAll()
	require.True(migrated)
	require.NoError(err)
}

func TestMigrateAllNoItem(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	migrated, err := kb.MigrateAll()
	require.False(migrated)
	require.NoError(err)
}

func TestMigrateErrUnknownItemKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey()

	legacyOfflineInfo := NewLegacyOfflineInfo(n1, pub, hd.Secp256k1.Name())
	serializedLegacyOfflineInfo := MarshalInfo(legacyOfflineInfo)

	item := keyring.Item{
		Key:         n1,
		Data:        serializedLegacyOfflineInfo,
		Description: "SDK kerying version",
	}

	ks, ok := kb.(keystore)
	require.True(ok)
	require.NoError(ks.setItem(item))

	incorrectItemKey := n1 + "1"
	migrated, err := ks.migrate(incorrectItemKey)
	require.False(migrated)
	require.EqualError(err, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, incorrectItemKey).Error())
}

func TestMigrateErrEmptyItemData(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	cdc := getCodec()

	require := require.New(t)
	kb, err := New(n1, BackendTest, dir, mockIn, cdc)
	require.NoError(err)

	item := keyring.Item{
		Key:         n1,
		Data:        []byte{},
		Description: "SDK kerying version",
	}

	ks, ok := kb.(keystore)
	require.True(ok)
	require.NoError(ks.setItem(item))

	migrated, err := ks.migrate(n1)
	require.False(migrated)
	require.EqualError(err, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, n1).Error())
}
