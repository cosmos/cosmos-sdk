package keyring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

const someKey = "theKey"

var unsupportedAlgo = AltSigningAlgo{
	Name:      "unsupported",
	DeriveKey: nil,
}

func TestAltKeyring_List(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	list, err := keyring.List()
	require.NoError(t, err)
	require.Empty(t, list)

	// Fails on creating unsupported SigningAlgo
	_, _, err = keyring.NewMnemonic("failing", English, unsupportedAlgo)
	require.EqualError(t, err, ErrUnsupportedSigningAlgo.Error())

	// Create 3 keys
	uid1, uid2, uid3 := "Zkey", "Bkey", "Rkey"
	_, _, err = keyring.NewMnemonic(uid1, English, AltSecp256k1)
	require.NoError(t, err)
	_, _, err = keyring.NewMnemonic(uid2, English, AltSecp256k1)
	require.NoError(t, err)
	_, _, err = keyring.NewMnemonic(uid3, English, AltSecp256k1)
	require.NoError(t, err)

	list, err = keyring.List()
	require.NoError(t, err)
	require.Len(t, list, 3)

	// Check they are in alphabetical order
	require.Equal(t, uid2, list[0].GetName())
	require.Equal(t, uid3, list[1].GetName())
	require.Equal(t, uid1, list[2].GetName())
}

func TestAltKeyring_NewAccount(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	entropy, err := bip39.NewEntropy(defaultEntropySize)
	require.NoError(t, err)

	mnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	theUid := "newUid"

	// Fails on creating unsupported SigningAlgo
	info, err := keyring.NewAccount(theUid, mnemonic, DefaultBIP39Passphrase, types.GetConfig().GetFullFundraiserPath(), unsupportedAlgo)
	require.EqualError(t, err, ErrUnsupportedSigningAlgo.Error())

	info, err = keyring.NewAccount(theUid, mnemonic, DefaultBIP39Passphrase, types.GetConfig().GetFullFundraiserPath(), AltSecp256k1)
	require.NoError(t, err)

	require.Equal(t, theUid, info.GetName())

	list, err := keyring.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestAltKeyring_SaveLedgerKey(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	// Test unsupported Algo
	i1, err := keyring.SaveLedgerKey("key", unsupportedAlgo, "cosmos", 0, 0)
	require.EqualError(t, err, ErrUnsupportedSigningAlgo.Error())

	i1, err = keyring.SaveLedgerKey("key", AltSecp256k1, "cosmos", 0, 0)
	if err != nil {
		require.Equal(t, "ledger nano S: support for ledger devices is not available in this executable", err.Error())
		t.Skip("ledger nano S: support for ledger devices is not available in this executable")
		return
	}
	require.Equal(t, "key", i1.GetName())
}

func TestAltKeyring_Get(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := someKey
	mnemonic, _, err := keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	key, err := keyring.Key(uid)
	require.NoError(t, err)
	requireEqualInfo(t, mnemonic, key)
}

func TestAltKeyring_KeyByAddress(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := someKey
	mnemonic, _, err := keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	key, err := keyring.KeyByAddress(mnemonic.GetAddress())
	require.NoError(t, err)
	requireEqualInfo(t, key, mnemonic)
}

func TestAltKeyring_Delete(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := someKey
	_, _, err = keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	list, err := keyring.List()
	require.NoError(t, err)
	require.Len(t, list, 1)

	err = keyring.Delete(uid)
	require.NoError(t, err)

	list, err = keyring.List()
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestAltKeyring_DeleteByAddress(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := someKey
	mnemonic, _, err := keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	list, err := keyring.List()
	require.NoError(t, err)
	require.Len(t, list, 1)

	err = keyring.DeleteByAddress(mnemonic.GetAddress())
	require.NoError(t, err)

	list, err = keyring.List()
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestAltKeyring_SavePubKey(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	list, err := keyring.List()
	require.NoError(t, err)
	require.Empty(t, list)

	key := "offline"
	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()

	info, err := keyring.SavePubKey(key, pub, AltSecp256k1)
	require.Nil(t, err)
	require.Equal(t, pub, info.GetPubKey())
	require.Equal(t, key, info.GetName())
	require.Equal(t, AltSecp256k1.Name, info.GetAlgo())

	list, err = keyring.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(list))
}

func TestAltKeyring_SaveMultisig(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	mnemonic1, _, err := keyring.NewMnemonic("key1", English, AltSecp256k1)
	require.NoError(t, err)
	mnemonic2, _, err := keyring.NewMnemonic("key2", English, AltSecp256k1)
	require.NoError(t, err)

	key := "multi"
	pub := multisig.NewPubKeyMultisigThreshold(2, []crypto.PubKey{mnemonic1.GetPubKey(), mnemonic2.GetPubKey()})

	info, err := keyring.SaveMultisig(key, pub)
	require.Nil(t, err)
	require.Equal(t, pub, info.GetPubKey())
	require.Equal(t, key, info.GetName())

	list, err := keyring.List()
	require.NoError(t, err)
	require.Len(t, list, 3)
}

func TestAltKeyring_Sign(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "jack"
	_, _, err = keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	msg := []byte("some message")

	sign, key, err := keyring.Sign(uid, msg)
	require.NoError(t, err)

	require.True(t, key.VerifyBytes(msg, sign))
}

func TestAltKeyring_SignByAddress(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "jack"
	mnemonic, _, err := keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	msg := []byte("some message")

	sign, key, err := keyring.SignByAddress(mnemonic.GetAddress(), msg)
	require.NoError(t, err)

	require.True(t, key.VerifyBytes(msg, sign))
}

func TestAltKeyring_ImportExportPrivKey(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "theId"
	_, _, err = keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	passphrase := "somePass"
	armor, err := keyring.ExportPrivKeyArmor(uid, passphrase)
	require.NoError(t, err)

	// Should fail importing private key on existing key.
	err = keyring.ImportPrivKey(uid, armor, passphrase)
	require.EqualError(t, err, fmt.Sprintf("cannot overwrite key: %s", uid))

	newUid := "theNewId"
	// Should fail importing with wrong password
	err = keyring.ImportPrivKey(newUid, armor, "wrongPass")
	require.EqualError(t, err, "failed to decrypt private key: ciphertext decryption failed")

	err = keyring.ImportPrivKey(newUid, armor, passphrase)
	assert.NoError(t, err)
}

func TestAltKeyring_ImportExportPrivKey_ByAddress(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "theId"
	mnemonic, _, err := keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	passphrase := "somePass"
	armor, err := keyring.ExportPrivKeyArmorByAddress(mnemonic.GetAddress(), passphrase)
	require.NoError(t, err)

	// Should fail importing private key on existing key.
	err = keyring.ImportPrivKey(uid, armor, passphrase)
	require.EqualError(t, err, fmt.Sprintf("cannot overwrite key: %s", uid))

	newUid := "theNewId"
	// Should fail importing with wrong password
	err = keyring.ImportPrivKey(newUid, armor, "wrongPass")
	require.EqualError(t, err, "failed to decrypt private key: ciphertext decryption failed")

	err = keyring.ImportPrivKey(newUid, armor, passphrase)
	assert.NoError(t, err)
}

func TestAltKeyring_ImportExportPubKey(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "theId"
	_, _, err = keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	armor, err := keyring.ExportPubKeyArmor(uid)
	require.NoError(t, err)

	// Should fail importing private key on existing key.
	err = keyring.ImportPubKey(uid, armor)
	require.EqualError(t, err, fmt.Sprintf("cannot overwrite data for name: %s", uid))

	newUid := "theNewId"
	err = keyring.ImportPubKey(newUid, armor)
	assert.NoError(t, err)
}

func TestAltKeyring_ImportExportPubKey_ByAddress(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "theId"
	mnemonic, _, err := keyring.NewMnemonic(uid, English, AltSecp256k1)
	require.NoError(t, err)

	armor, err := keyring.ExportPubKeyArmorByAddress(mnemonic.GetAddress())
	require.NoError(t, err)

	// Should fail importing private key on existing key.
	err = keyring.ImportPubKey(uid, armor)
	require.EqualError(t, err, fmt.Sprintf("cannot overwrite data for name: %s", uid))

	newUid := "theNewId"
	err = keyring.ImportPubKey(newUid, armor)
	assert.NoError(t, err)
}

func requireEqualInfo(t *testing.T, key Info, mnemonic Info) {
	require.Equal(t, key.GetName(), mnemonic.GetName())
	require.Equal(t, key.GetAddress(), mnemonic.GetAddress())
	require.Equal(t, key.GetPubKey(), mnemonic.GetPubKey())
	require.Equal(t, key.GetAlgo(), mnemonic.GetAlgo())
	require.Equal(t, key.GetType(), mnemonic.GetType())
}
