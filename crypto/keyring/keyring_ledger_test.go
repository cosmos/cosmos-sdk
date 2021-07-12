//+build ledger test_ledger_mock

package keyring_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
)

func TestInMemoryCreateLedger(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	kb := keyring.NewInMemory(encCfg.Codec)

	k, err := kb.SaveLedgerKey("some_account", hd.Secp256k1, "cosmos", 118, 3, 1)

	if err != nil {
		require.Error(t, err)
		require.Equal(t, "ledger nano S: support for ledger devices is not available in this executable", err.Error())
		require.Nil(t, k)
		t.Skip("ledger nano S: support for ledger devices is not available in this executable")
		return
	}

	// The mock is available, check that the address is correct
	pubKey, err := k.GetPubKey()
	require.NoError(t, err)
	expectedPkStr := "PubKeySecp256k1{03602C0CB4D8C0081FEE794BDE96E7B95FA16F2B5283B764AC070584327B2C7202}"
	require.Equal(t, expectedPkStr, pubKey.String())

	// Check that restoring the key gets the same results
	restoredRecord, err := kb.Key("some_account")
	require.NoError(t, err)
	require.NotNil(t, restoredRecord)
	require.Equal(t, "some_account", restoredRecord.Name)
	// TODO handle GetType remove it
	//require.Equal(t, TypeLedger, restoredKey.GetType())
	pubKey, err = restoredRecord.GetPubKey()
	require.NoError(t, err)
	require.Equal(t, expectedPkStr, pubKey.String())

	ledgerInfo := restoredRecord.GetLedger()
	require.NotNil(t, ledgerInfo)
	path := ledgerInfo.GetPath()
	require.Equal(t, "m/44'/118'/3'/0/1", path.String())
}

// TestSignVerify does some detailed checks on how we sign and validate
// signatures
func TestSignVerifyKeyRingWithLedger(t *testing.T) {
	dir := t.TempDir()
	encCfg := simapp.MakeTestEncodingConfig()

	kb, err := keyring.New("keybasename", "test", dir, nil, encCfg.Codec)
	require.NoError(t, err)

	k, err := kb.SaveLedgerKey("key", hd.Secp256k1, "cosmos", 118, 0, 0)
	if err != nil {
		require.Equal(t, "ledger nano S: support for ledger devices is not available in this executable", err.Error())
		t.Skip("ledger nano S: support for ledger devices is not available in this executable")
		return
	}
	require.Equal(t, "key", k.Name)

	d1 := []byte("my first message")
	s1, pub1, err := kb.Sign("key", d1)
	require.NoError(t, err)

	s2, pub2, err := keyring.SignWithLedger(k, d1)
	require.NoError(t, err)

	require.True(t, pub1.Equals(pub2))
	require.True(t, bytes.Equal(s1, s2))

	key1, err := k.GetPubKey()
	require.NoError(t, err)

	require.Equal(t, key1, pub1)
	require.Equal(t, key1, pub2)
	require.True(t, pub1.VerifySignature(d1, s1))
	require.True(t, key1.VerifySignature(d1, s1))
	require.True(t, bytes.Equal(s1, s2))

	k, _, err = kb.NewMnemonic("test", keyring.English, types.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	_, _, err = keyring.SignWithLedger(k, d1)
	require.Error(t, err)
	require.Equal(t, "not a ledger object", err.Error())
}

func TestAltKeyring_SaveLedgerKey(t *testing.T) {
	dir := t.TempDir()
	encCfg := simapp.MakeTestEncodingConfig()

	kr, err := keyring.New(t.Name(), keyring.BackendTest, dir, nil, encCfg.Codec)
	require.NoError(t, err)

	// Test unsupported Algo
	_, err = kr.SaveLedgerKey("key", keyring.NotSupportedAlgo{}, "cosmos", 118, 0, 0)
	require.EqualError(t, err, keyring.ErrUnsupportedSigningAlgo.Error())

	k, err := kr.SaveLedgerKey("some_account", hd.Secp256k1, "cosmos", 118, 3, 1)
	if err != nil {
		require.Equal(t, "ledger nano S: support for ledger devices is not available in this executable", err.Error())
		t.Skip("ledger nano S: support for ledger devices is not available in this executable")
		return
	}
	// The mock is available, check that the address is correct
	require.Equal(t, "some_account", k.Name)
	pubKey, err := k.GetPubKey()
	require.NoError(t, err)
	expectedPkStr := "PubKeySecp256k1{03602C0CB4D8C0081FEE794BDE96E7B95FA16F2B5283B764AC070584327B2C7202}"
	require.Equal(t, expectedPkStr, pubKey.String())

	// Check that restoring the key gets the same results
	restoredRecord, err := kr.Key("some_account")
	require.NoError(t, err)
	require.NotNil(t, restoredRecord)
	require.Equal(t, "some_account", restoredRecord.Name)
	//	require.Equal(t, TypeLedger, restoredRecord.GetType())
	pubKey, err = restoredRecord.GetPubKey()
	require.NoError(t, err)
	require.Equal(t, expectedPkStr, pubKey.String())

	ledgerInfo := k.GetLedger()
	require.NotNil(t, ledgerInfo)

	path := ledgerInfo.GetPath()
	require.Equal(t, "m/44'/118'/3'/0/1", path.String())
}
