package keyring_test

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/stretchr/testify/require"
	sdk "github.com/cosmos/cosmos-sdk/types"
	
)

/* cases

3)create radnomBytes -> err

*/

// create legacyInfo -> keyring -> checkMigrate -> key will be migrated
// TODO for legacyLedger , legacyOffline and legacyMulti table driven test
func TestMigrationOneLegacyLocalKey(t *testing.T) {

	const n1 = "cosmos"
	
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Marshaler)
	require.NoError(err)
	
	//saves legacyLocalInfo to keyring 
	_,_, err = kb.NewLegacyMnemonic(n1, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)
	
	//calls checkMigrate, migrates the key to proto
	migrated, err := kb.CheckMigrate()
	require.True(migrated)
	require.NoError(err)
}

//create Record - no migration required make for all types of keys Local,Ledger etc
func TestMigrationRecord(t *testing.T) {

	const n1 = "cosmos"
	
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendTest, dir, mockIn, encCfg.Marshaler)
	require.NoError(err)
	
	_,_, err = kb.NewMnemonic(n1, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)
	
	migrated, err := kb.CheckMigrate()
	require.False(migrated)
	require.NoError(err)
}



