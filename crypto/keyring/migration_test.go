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

/*
create legacyInfo -> keyring -> checkMigrate -> will be migrated
create just record-> no migrate
create radnomBytes -> err

*/

func TestMigrationOneLegacyKey(t *testing.T) {

	const n1 = "cosmos"
	
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	encCfg := simapp.MakeTestEncodingConfig()

	require := require.New(t)
	kb, err := keyring.New(n1, keyring.BackendFile, dir, mockIn, encCfg.Marshaler)
	require.NoError(err)
	 
	_,_, err = kb.NewLegacyMnemonic(n1, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)

	k, err := kb.Key(n1)
	require.NotNil(k)
	require.NoError(err)
	require.Equal(k.Name, n1)

}
