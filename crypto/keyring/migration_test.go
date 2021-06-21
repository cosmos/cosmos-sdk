package keyring_test

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	dkeyring "github.com/99designs/keyring"
)

// TODO
// create radnomBytes -> err test
// create table driven tests

// create legacyInfo -> keyring -> checkMigrate -> key will be migrated

const n1 = "cosmos"

func TestMigrationLegacyLocalKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Marshaler)
	require.NoError(err)
	
	//saves legacyLocalInfo to keyring 
	accountType := "local"
	info,_, err := kb.NewLegacyMnemonic(n1, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1, accountType)
	require.NoError(err)
	require.Equal(info.GetName(),n1)
	
	//calls checkMigrate, migrates the amino key to proto
	migrated, err := kb.CheckMigrate()
	require.True(migrated)
	require.NoError(err)
}

// TODO fix error support for ledger devices is not available in this executable
func TestMigrationLegacyLedgerKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Marshaler)
	require.NoError(err)
	
	//saves legacyLocalInfo to keyring 
	accountType := "ledger"
	info,_, err := kb.NewLegacyMnemonic(n1, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1, accountType)
	require.NoError(err)
	require.Equal(info.GetName(), n1)
	
	//calls checkMigrate, migrates the key to proto
	migrated, err := kb.CheckMigrate()
	require.True(migrated)
	require.NoError(err)
}

func TestMigrationLegacyOfflineKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Marshaler)
	require.NoError(err)
	
	//saves legacyLocalInfo to keyring 
	accountType := "offline"
	info,_, err := kb.NewLegacyMnemonic(n1, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1, accountType)
	require.NoError(err)
	require.Equal(info.GetName(), n1)
	
	//calls checkMigrate, migrates the key to proto
	migrated, err := kb.CheckMigrate()
	require.True(migrated)
	require.NoError(err)
}

func TestMigrationLegacyMultiKey(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Marshaler)
	require.NoError(err)
	
	//saves legacyLocalInfo to keyring 
	accountType :=  "multi"
	info,_, err := kb.NewLegacyMnemonic(n1, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1, accountType)
	require.NoError(err)
	require.Equal(info.GetName(), n1)
	
	//calls checkMigrate, migrates the key to proto
	migrated, err := kb.CheckMigrate()
	require.True(migrated)
	require.NoError(err)
}

// TODO  do i need to test migration for ledger,offline record items as well?
func TestMigrationLocalRecord(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Marshaler)
	require.NoError(err)
	
	k,_, err := kb.NewMnemonic(n1, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)
	require.Equal(k.Name, n1)
	
	migrated, err := kb.CheckMigrate()
	require.False(migrated)
	require.NoError(err)
}

// TODO insert multiple incorrect migration keys and output errors to user 
func TestMigrationOneRandomItemError(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Marshaler)
	require.NoError(err)

	randomBytes := []byte("abckd0s03l")

	errItem := dkeyring.Item{
		Key:         keyring.InfoKey(n1),
		Data:        randomBytes,
		Description: "SDK kerying version",
	}

	err = kb.SetItem(errItem)
	require.NoError(err)

	migrated, err := kb.CheckMigrate()
	require.False(migrated)
	require.Error(err)
}






