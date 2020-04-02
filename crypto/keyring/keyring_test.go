package keyring

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func init() {
	crypto.BcryptSecurityParameter = 1
}

const (
	nums   = "1234"
	foobar = "foobar"
)

func TestNewKeyring(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	mockIn := strings.NewReader("")
	t.Cleanup(cleanup)
	kr, err := NewKeyring("cosmos", BackendFile, dir, mockIn)
	require.NoError(t, err)

	mockIn.Reset("password\npassword\n")
	info, _, err := kr.CreateMnemonic("foo", English, "password", Secp256k1)
	require.NoError(t, err)
	require.Equal(t, "foo", info.GetName())
}

func TestKeyManagementKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	algo := Secp256k1
	n1, n2, n3 := "personal", "business", "other"
	p1, p2 := "1234", "really-secure!@#$"

	// Check empty state
	l, err := kb.List()
	require.Nil(t, err)
	assert.Empty(t, l)

	_, _, err = kb.CreateMnemonic(n1, English, p1, Ed25519)
	require.Error(t, err, "ed25519 keys are currently not supported by keybase")

	// create some keys
	_, err = kb.Get(n1)
	require.Error(t, err)
	i, _, err := kb.CreateMnemonic(n1, English, p1, algo)

	require.NoError(t, err)
	require.Equal(t, n1, i.GetName())
	_, _, err = kb.CreateMnemonic(n2, English, p2, algo)
	require.NoError(t, err)

	// we can get these keys
	i2, err := kb.Get(n2)
	require.NoError(t, err)
	_, err = kb.Get(n3)
	require.NotNil(t, err)
	_, err = kb.GetByAddress(accAddr(i2))
	require.NoError(t, err)
	addr, err := sdk.AccAddressFromBech32("cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t")
	require.NoError(t, err)
	_, err = kb.GetByAddress(addr)
	require.NotNil(t, err)

	// list shows them in order
	keyS, err := kb.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))
	// note these are in alphabetical order
	require.Equal(t, n2, keyS[0].GetName())
	require.Equal(t, n1, keyS[1].GetName())
	require.Equal(t, i2.GetPubKey(), keyS[0].GetPubKey())

	// deleting a key removes it
	err = kb.Delete("bad name", "foo", false)
	require.NotNil(t, err)
	err = kb.Delete(n1, p1, false)
	require.NoError(t, err)
	keyS, err = kb.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))
	_, err = kb.Get(n1)
	require.Error(t, err)

	// create an offline key
	o1 := "offline"
	priv1 := ed25519.GenPrivKey()
	pub1 := priv1.PubKey()
	i, err = kb.CreateOffline(o1, pub1, Ed25519)
	require.Nil(t, err)
	require.Equal(t, pub1, i.GetPubKey())
	require.Equal(t, o1, i.GetName())
	keyS, err = kb.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))

	// delete the offline key
	err = kb.Delete(o1, "", false)
	require.NoError(t, err)
	keyS, err = kb.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))

	// addr cache gets nuked - and test skip flag
	require.NoError(t, kb.Delete(n2, "", true))
}

// TestSignVerify does some detailed checks on how we sign and validate
// signatures
func TestSignVerifyKeyRingWithLedger(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	i1, err := kb.CreateLedger("key", Secp256k1, "cosmos", 0, 0)
	if err != nil {
		require.Equal(t, "ledger nano S: support for ledger devices is not available in this executable", err.Error())
		t.Skip("ledger nano S: support for ledger devices is not available in this executable")
		return
	}
	require.Equal(t, "key", i1.GetName())

	p1 := "1234"
	d1 := []byte("my first message")
	s1, pub1, err := kb.Sign("key", p1, d1)
	require.NoError(t, err)

	s2, pub2, err := SignWithLedger(i1, d1)
	require.NoError(t, err)

	require.Equal(t, i1.GetPubKey(), pub1)
	require.Equal(t, i1.GetPubKey(), pub2)
	require.True(t, pub1.VerifyBytes(d1, s1))
	require.True(t, i1.GetPubKey().VerifyBytes(d1, s1))
	require.True(t, bytes.Equal(s1, s2))

	localInfo, _, err := kb.CreateMnemonic("test", English, p1, Secp256k1)
	require.NoError(t, err)
	_, _, err = SignWithLedger(localInfo, d1)
	require.Error(t, err)
	require.Equal(t, "not a ledger object", err.Error())
}

func TestSignVerifyKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)
	algo := Secp256k1

	n1, n2, n3 := "some dude", "a dudette", "dude-ish"
	p1, p2, p3 := "1234", "foobar", "foobar"

	// create two users and get their info
	i1, _, err := kb.CreateMnemonic(n1, English, p1, algo)
	require.Nil(t, err)

	i2, _, err := kb.CreateMnemonic(n2, English, p2, algo)
	require.Nil(t, err)

	// Import a public key
	armor, err := kb.ExportPubKey(n2)
	require.Nil(t, err)
	err = kb.ImportPubKey(n3, armor)
	require.NoError(t, err)
	i3, err := kb.Get(n3)
	require.NoError(t, err)
	require.Equal(t, i3.GetName(), n3)

	// let's try to sign some messages
	d1 := []byte("my first message")
	d2 := []byte("some other important info!")
	d3 := []byte("feels like I forgot something...")

	// try signing both data with both ..
	s11, pub1, err := kb.Sign(n1, p1, d1)
	require.Nil(t, err)
	require.Equal(t, i1.GetPubKey(), pub1)

	s12, pub1, err := kb.Sign(n1, p1, d2)
	require.Nil(t, err)
	require.Equal(t, i1.GetPubKey(), pub1)

	s21, pub2, err := kb.Sign(n2, p2, d1)
	require.Nil(t, err)
	require.Equal(t, i2.GetPubKey(), pub2)

	s22, pub2, err := kb.Sign(n2, p2, d2)
	require.Nil(t, err)
	require.Equal(t, i2.GetPubKey(), pub2)

	// let's try to validate and make sure it only works when everything is proper
	cases := []struct {
		key   tmcrypto.PubKey
		data  []byte
		sig   []byte
		valid bool
	}{
		// proper matches
		{i1.GetPubKey(), d1, s11, true},
		// change data, pubkey, or signature leads to fail
		{i1.GetPubKey(), d2, s11, false},
		{i2.GetPubKey(), d1, s11, false},
		{i1.GetPubKey(), d1, s21, false},
		// make sure other successes
		{i1.GetPubKey(), d2, s12, true},
		{i2.GetPubKey(), d1, s21, true},
		{i2.GetPubKey(), d2, s22, true},
	}

	for i, tc := range cases {
		valid := tc.key.VerifyBytes(tc.data, tc.sig)
		require.Equal(t, tc.valid, valid, "%d", i)
	}

	// Now try to sign data with a secret-less key
	_, _, err = kb.Sign(n3, p3, d3)
	require.Error(t, err)
	require.Equal(t, "cannot sign with offline keys", err.Error())
}

func TestExportImportKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	info, _, err := kb.CreateMnemonic("john", English, "secretcpw", Secp256k1)
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")

	john, err := kb.Get("john")
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")
	johnAddr := info.GetPubKey().Address()

	armor, err := kb.Export("john")
	require.NoError(t, err)

	err = kb.Import("john2", armor)
	require.NoError(t, err)

	john2, err := kb.Get("john2")
	require.NoError(t, err)

	require.Equal(t, john.GetPubKey().Address(), johnAddr)
	require.Equal(t, john.GetName(), "john")
	require.Equal(t, john, john2)
}

func TestExportImportPubKeyKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)
	algo := Secp256k1

	// CreateMnemonic a private-public key pair and ensure consistency
	notPasswd := "n9y25ah7"
	info, _, err := kb.CreateMnemonic("john", English, notPasswd, algo)
	require.Nil(t, err)
	require.NotEqual(t, info, "")
	require.Equal(t, info.GetName(), "john")
	addr := info.GetPubKey().Address()
	john, err := kb.Get("john")
	require.NoError(t, err)
	require.Equal(t, john.GetName(), "john")
	require.Equal(t, john.GetPubKey().Address(), addr)

	// Export the public key only
	armor, err := kb.ExportPubKey("john")
	require.NoError(t, err)
	// Import it under a different name
	err = kb.ImportPubKey("john-pubkey-only", armor)
	require.NoError(t, err)
	// Ensure consistency
	john2, err := kb.Get("john-pubkey-only")
	require.NoError(t, err)
	// Compare the public keys
	require.True(t, john.GetPubKey().Equals(john2.GetPubKey()))
	// Ensure the original key hasn't changed
	john, err = kb.Get("john")
	require.NoError(t, err)
	require.Equal(t, john.GetPubKey().Address(), addr)
	require.Equal(t, john.GetName(), "john")

	// Ensure keys cannot be overwritten
	err = kb.ImportPubKey("john-pubkey-only", armor)
	require.NotNil(t, err)
}

func TestExportPrivateKeyObjectKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	info, _, err := kb.CreateMnemonic("john", English, "secretcpw", Secp256k1)
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")

	// export private key object
	exported, err := kb.ExportPrivateKeyObject("john", "secretcpw")
	require.Nil(t, err, "%+v", err)
	require.True(t, exported.PubKey().Equals(info.GetPubKey()))
}

func TestAdvancedKeyManagementKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	algo := Secp256k1
	n1, n2 := "old-name", "new name"
	p1 := "1234"

	// make sure key works with initial password
	_, _, err = kb.CreateMnemonic(n1, English, p1, algo)
	require.Nil(t, err, "%+v", err)

	_, err = kb.Export(n1 + ".notreal")
	require.NotNil(t, err)
	_, err = kb.Export(" " + n1)
	require.NotNil(t, err)
	_, err = kb.Export(n1 + " ")
	require.NotNil(t, err)
	_, err = kb.Export("")
	require.NotNil(t, err)
	exported, err := kb.Export(n1)
	require.Nil(t, err, "%+v", err)

	// import succeeds
	err = kb.Import(n2, exported)
	require.NoError(t, err)

	// second import fails
	err = kb.Import(n2, exported)
	require.NotNil(t, err)
}

func TestSeedPhraseKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	algo := Secp256k1
	n1, n2 := "lost-key", "found-again"
	p1, p2 := "1234", "foobar"

	// make sure key works with initial password
	info, mnemonic, err := kb.CreateMnemonic(n1, English, p1, algo)
	require.Nil(t, err, "%+v", err)
	require.Equal(t, n1, info.GetName())
	assert.NotEmpty(t, mnemonic)

	// now, let us delete this key
	err = kb.Delete(n1, p1, false)
	require.Nil(t, err, "%+v", err)
	_, err = kb.Get(n1)
	require.NotNil(t, err)

	// let us re-create it from the mnemonic-phrase
	params := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	hdPath := params.String()
	newInfo, err := kb.CreateAccount(n2, mnemonic, DefaultBIP39Passphrase, p2, hdPath, Secp256k1)
	require.NoError(t, err)
	require.Equal(t, n2, newInfo.GetName())
	require.Equal(t, info.GetPubKey().Address(), newInfo.GetPubKey().Address())
	require.Equal(t, info.GetPubKey(), newInfo.GetPubKey())
}

func TestKeyringKeybaseExportImportPrivKey(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)
	_, _, err = kb.CreateMnemonic("john", English, "password", Secp256k1)
	require.NoError(t, err)

	// no error, password is irrelevant, keystr cointains ASCII armored private key
	keystr, err := kb.ExportPrivKey("john", "wrongpassword", "password")
	require.NoError(t, err)
	require.NotEmpty(t, keystr)

	// try import the key - wrong password
	err = kb.ImportPrivKey("john2", keystr, "somepassword")
	require.Equal(t, "failed to decrypt private key: ciphertext decryption failed", err.Error())

	// try import the key with the correct password
	require.NoError(t, kb.ImportPrivKey("john2", keystr, "password"))

	// overwrite is not allowed
	err = kb.ImportPrivKey("john2", keystr, "password")
	require.Equal(t, "cannot overwrite key: john2", err.Error())

	// try export non existing key
	_, err = kb.ExportPrivKey("john3", "wrongpassword", "password")
	require.Equal(t, "The specified item could not be found in the keyring", err.Error())
}

func TestSupportedAlgos(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)
	require.Equal(t, []SigningAlgo{"secp256k1"}, kb.SupportedAlgos())
	require.Equal(t, []SigningAlgo{"secp256k1"}, kb.SupportedAlgosLedger())
}

func TestCustomDerivFuncKey(t *testing.T) {
	kb := NewInMemory(WithDeriveFunc(func(mnemonic string, bip39Passphrase, hdPath string, algo SigningAlgo) ([]byte, error) {
		return nil, errors.New("cannot derive keys")
	}))
	_, _, err := kb.CreateMnemonic("test", English, "", "")
	require.Error(t, err, "cannot derive keys")
}

func TestInMemoryLanguage(t *testing.T) {
	kb := NewInMemory()
	_, _, err := kb.CreateMnemonic("something", Japanese, "no_pass", Secp256k1)
	require.Error(t, err)
	require.Equal(t, "unsupported language: only english is supported", err.Error())
}

func TestInMemoryCreateMultisig(t *testing.T) {
	kb, err := NewKeyring("keybasename", "memory", "", nil)
	require.NoError(t, err)
	multi := multisig.PubKeyMultisigThreshold{
		K:       1,
		PubKeys: []tmcrypto.PubKey{secp256k1.GenPrivKey().PubKey()},
	}
	_, err = kb.CreateMulti("multi", multi)
	require.NoError(t, err)
}

func TestInMemoryCreateAccountInvalidMnemonic(t *testing.T) {
	kb := NewInMemory()
	_, err := kb.CreateAccount(
		"some_account",
		"malarkey pair crucial catch public canyon evil outer stage ten gym tornado",
		"", "", CreateHDPath(0, 0).String(), Secp256k1)
	require.Error(t, err)
	require.Equal(t, "Invalid mnemonic", err.Error())
}

func TestInMemoryCreateLedgerUnsupportedAlgo(t *testing.T) {
	kb := NewInMemory()

	supportedLedgerAlgos := kb.SupportedAlgosLedger()
	for _, supportedAlgo := range supportedLedgerAlgos {
		if Ed25519 == supportedAlgo {
			require.FailNow(t, "Was not an unsupported algorithm")
		}
	}

	_, err := kb.CreateLedger("some_account", Ed25519, "cosmos", 0, 1)
	require.Error(t, err)
	require.Equal(t, "unsupported signing algo", err.Error())
}

func TestInMemoryCreateLedger(t *testing.T) {
	kb := NewInMemory(WithSupportedAlgosLedger([]SigningAlgo{Secp256k1, Ed25519}))

	// test_cover and test_unit will result in different answers
	// test_cover does not compile some dependencies so ledger is disabled
	// test_unit may add a ledger mock
	// both cases are acceptable
	supportedLedgerAlgos := kb.SupportedAlgosLedger()
	secpSupported := false
	edSupported := false
	for _, supportedAlgo := range supportedLedgerAlgos {
		secpSupported = secpSupported || (supportedAlgo == Secp256k1)
		edSupported = edSupported || (supportedAlgo == Ed25519)
	}
	require.True(t, secpSupported)
	require.True(t, edSupported)

	ledger, err := kb.CreateLedger("some_account", Secp256k1, "cosmos", 3, 1)

	if err != nil {
		require.Error(t, err)
		require.Equal(t, "ledger nano S: support for ledger devices is not available in this executable", err.Error())
		require.Nil(t, ledger)
		t.Skip("ledger nano S: support for ledger devices is not available in this executable")
		return
	}

	// The mock is available, check that the address is correct
	pubKey := ledger.GetPubKey()
	pk, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubKey)
	require.NoError(t, err)
	require.Equal(t, "cosmospub1addwnpepqdszcr95mrqqs8lw099aa9h8h906zmet22pmwe9vquzcgvnm93eqygufdlv", pk)

	// Check that restoring the key gets the same results
	restoredKey, err := kb.Get("some_account")
	require.NoError(t, err)
	require.NotNil(t, restoredKey)
	require.Equal(t, "some_account", restoredKey.GetName())
	require.Equal(t, TypeLedger, restoredKey.GetType())
	pubKey = restoredKey.GetPubKey()
	pk, err = sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubKey)
	require.NoError(t, err)
	require.Equal(t, "cosmospub1addwnpepqdszcr95mrqqs8lw099aa9h8h906zmet22pmwe9vquzcgvnm93eqygufdlv", pk)

	path, err := restoredKey.GetPath()
	require.NoError(t, err)
	require.Equal(t, "44'/118'/3'/0/1", path.String())
}

// TestInMemoryKeyManagement makes sure we can manipulate these keys well
func TestInMemoryKeyManagement(t *testing.T) {
	// make the storage with reasonable defaults
	cstore := NewInMemory(WithSupportedAlgos([]SigningAlgo{Secp256k1, Sr25519}))

	// Test modified supported algos
	supportedAlgos := cstore.SupportedAlgos()
	secpSupported := false
	edSupported := false
	srSupported := false
	for _, supportedAlgo := range supportedAlgos {
		secpSupported = secpSupported || (supportedAlgo == Secp256k1)
		edSupported = edSupported || (supportedAlgo == Ed25519)
		srSupported = srSupported || (supportedAlgo == Sr25519)
	}
	require.True(t, secpSupported)
	require.False(t, edSupported)
	require.True(t, srSupported)

	algo := Secp256k1
	n1, n2, n3 := "personal", "business", "other"
	p1, p2 := nums, "really-secure!@#$"

	// Check empty state
	l, err := cstore.List()
	require.Nil(t, err)
	require.Empty(t, l)

	_, _, err = cstore.CreateMnemonic(n1, English, p1, Ed25519)
	require.Error(t, err, "ed25519 keys are currently not supported by keybase")

	// create some keys
	_, err = cstore.Get(n1)
	require.Error(t, err)
	i, _, err := cstore.CreateMnemonic(n1, English, p1, algo)

	require.NoError(t, err)
	require.Equal(t, n1, i.GetName())
	_, _, err = cstore.CreateMnemonic(n2, English, p2, algo)
	require.NoError(t, err)

	// we can get these keys
	i2, err := cstore.Get(n2)
	require.NoError(t, err)
	_, err = cstore.Get(n3)
	require.NotNil(t, err)
	_, err = cstore.GetByAddress(accAddr(i2))
	require.NoError(t, err)
	addr, err := sdk.AccAddressFromBech32("cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t")
	require.NoError(t, err)
	_, err = cstore.GetByAddress(addr)
	require.NotNil(t, err)

	// list shows them in order
	keyS, err := cstore.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))
	// note these are in alphabetical order
	require.Equal(t, n2, keyS[0].GetName())
	require.Equal(t, n1, keyS[1].GetName())
	require.Equal(t, i2.GetPubKey(), keyS[0].GetPubKey())

	// deleting a key removes it
	err = cstore.Delete("bad name", "foo", false)
	require.NotNil(t, err)
	err = cstore.Delete(n1, p1, false)
	require.NoError(t, err)
	keyS, err = cstore.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))
	_, err = cstore.Get(n1)
	require.Error(t, err)

	// create an offline key
	o1 := "offline"
	priv1 := ed25519.GenPrivKey()
	pub1 := priv1.PubKey()
	i, err = cstore.CreateOffline(o1, pub1, algo)
	require.Nil(t, err)
	require.Equal(t, pub1, i.GetPubKey())
	require.Equal(t, o1, i.GetName())
	iOffline := i.(*offlineInfo)
	require.Equal(t, algo, iOffline.GetAlgo())
	keyS, err = cstore.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))

	// delete the offline key
	err = cstore.Delete(o1, "", false)
	require.NoError(t, err)
	keyS, err = cstore.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))

	// addr cache gets nuked - and test skip flag
	err = cstore.Delete(n2, "", true)
	require.NoError(t, err)
}

// TestInMemorySignVerify does some detailed checks on how we sign and validate
// signatures
func TestInMemorySignVerify(t *testing.T) {
	cstore := NewInMemory()
	algo := Secp256k1

	n1, n2, n3 := "some dude", "a dudette", "dude-ish"
	p1, p2, p3 := nums, foobar, foobar

	// create two users and get their info
	i1, _, err := cstore.CreateMnemonic(n1, English, p1, algo)
	require.Nil(t, err)

	i2, _, err := cstore.CreateMnemonic(n2, English, p2, algo)
	require.Nil(t, err)

	// Import a public key
	armor, err := cstore.ExportPubKey(n2)
	require.Nil(t, err)
	err = cstore.ImportPubKey(n3, armor)
	require.NoError(t, err)
	i3, err := cstore.Get(n3)
	require.NoError(t, err)
	require.Equal(t, i3.GetName(), n3)

	// let's try to sign some messages
	d1 := []byte("my first message")
	d2 := []byte("some other important info!")
	d3 := []byte("feels like I forgot something...")

	// try signing both data with both ..
	s11, pub1, err := cstore.Sign(n1, p1, d1)
	require.Nil(t, err)
	require.Equal(t, i1.GetPubKey(), pub1)

	s12, pub1, err := cstore.Sign(n1, p1, d2)
	require.Nil(t, err)
	require.Equal(t, i1.GetPubKey(), pub1)

	s21, pub2, err := cstore.Sign(n2, p2, d1)
	require.Nil(t, err)
	require.Equal(t, i2.GetPubKey(), pub2)

	s22, pub2, err := cstore.Sign(n2, p2, d2)
	require.Nil(t, err)
	require.Equal(t, i2.GetPubKey(), pub2)

	// let's try to validate and make sure it only works when everything is proper
	cases := []struct {
		key   tmcrypto.PubKey
		data  []byte
		sig   []byte
		valid bool
	}{
		// proper matches
		{i1.GetPubKey(), d1, s11, true},
		// change data, pubkey, or signature leads to fail
		{i1.GetPubKey(), d2, s11, false},
		{i2.GetPubKey(), d1, s11, false},
		{i1.GetPubKey(), d1, s21, false},
		// make sure other successes
		{i1.GetPubKey(), d2, s12, true},
		{i2.GetPubKey(), d1, s21, true},
		{i2.GetPubKey(), d2, s22, true},
	}

	for i, tc := range cases {
		valid := tc.key.VerifyBytes(tc.data, tc.sig)
		require.Equal(t, tc.valid, valid, "%d", i)
	}

	// Now try to sign data with a secret-less key
	_, _, err = cstore.Sign(n3, p3, d3)
	require.Error(t, err)
	require.Equal(t, "cannot sign with offline keys", err.Error())
}

// TestInMemoryExportImport tests exporting and importing
func TestInMemoryExportImport(t *testing.T) {
	// make the storage with reasonable defaults
	cstore := NewInMemory()

	info, _, err := cstore.CreateMnemonic("john", English, "secretcpw", Secp256k1)
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")

	john, err := cstore.Get("john")
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")
	johnAddr := info.GetPubKey().Address()

	armor, err := cstore.Export("john")
	require.NoError(t, err)

	err = cstore.Import("john2", armor)
	require.NoError(t, err)

	john2, err := cstore.Get("john2")
	require.NoError(t, err)

	require.Equal(t, john.GetPubKey().Address(), johnAddr)
	require.Equal(t, john.GetName(), "john")
	require.Equal(t, john, john2)
}

func TestInMemoryExportImportPrivKey(t *testing.T) {
	kb := NewInMemory()

	info, _, err := kb.CreateMnemonic("john", English, "secretcpw", Secp256k1)
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")
	priv1, err := kb.Get("john")
	require.NoError(t, err)

	// decrypt local private key, and produce encrypted ASCII armored output
	armored, err := kb.ExportPrivKey("john", "secretcpw", "new_secretcpw")
	require.NoError(t, err)

	// delete exported key
	require.NoError(t, kb.Delete("john", "", true))
	_, err = kb.Get("john")
	require.Error(t, err)

	// import armored key
	require.NoError(t, kb.ImportPrivKey("john", armored, "new_secretcpw"))

	// ensure old and new keys match
	priv2, err := kb.Get("john")
	require.NoError(t, err)
	require.True(t, priv1.GetPubKey().Equals(priv2.GetPubKey()))
}

func TestInMemoryExportImportPubKey(t *testing.T) {
	// make the storage with reasonable defaults
	cstore := NewInMemory()

	// CreateMnemonic a private-public key pair and ensure consistency
	notPasswd := "n9y25ah7"
	info, _, err := cstore.CreateMnemonic("john", English, notPasswd, Secp256k1)
	require.Nil(t, err)
	require.NotEqual(t, info, "")
	require.Equal(t, info.GetName(), "john")
	addr := info.GetPubKey().Address()
	john, err := cstore.Get("john")
	require.NoError(t, err)
	require.Equal(t, john.GetName(), "john")
	require.Equal(t, john.GetPubKey().Address(), addr)

	// Export the public key only
	armor, err := cstore.ExportPubKey("john")
	require.NoError(t, err)
	// Import it under a different name
	err = cstore.ImportPubKey("john-pubkey-only", armor)
	require.NoError(t, err)
	// Ensure consistency
	john2, err := cstore.Get("john-pubkey-only")
	require.NoError(t, err)
	// Compare the public keys
	require.True(t, john.GetPubKey().Equals(john2.GetPubKey()))
	// Ensure the original key hasn't changed
	john, err = cstore.Get("john")
	require.NoError(t, err)
	require.Equal(t, john.GetPubKey().Address(), addr)
	require.Equal(t, john.GetName(), "john")

	// Ensure keys cannot be overwritten
	err = cstore.ImportPubKey("john-pubkey-only", armor)
	require.NotNil(t, err)
}

func TestInMemoryExportPrivateKeyObject(t *testing.T) {
	kb := NewInMemory()

	info, _, err := kb.CreateMnemonic("john", English, "secretcpw", Secp256k1)
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")

	// export private key object
	_, err = kb.ExportPrivateKeyObject("john", "invalid")
	require.NoError(t, err, "%+v", err)
	exported, err := kb.ExportPrivateKeyObject("john", "secretcpw")
	require.Nil(t, err, "%+v", err)
	require.True(t, exported.PubKey().Equals(info.GetPubKey()))
}

// TestInMemoryAdvancedKeyManagement verifies update, import, export functionality
func TestInMemoryAdvancedKeyManagement(t *testing.T) {
	// make the storage with reasonable defaults
	cstore := NewInMemory()

	algo := Secp256k1
	n1, n2 := "old-name", "new name"
	p1 := nums

	// make sure key works with initial password
	_, _, err := cstore.CreateMnemonic(n1, English, p1, algo)
	require.Nil(t, err, "%+v", err)

	// exporting requires the proper name and passphrase
	_, err = cstore.Export(n1 + ".notreal")
	require.NotNil(t, err)
	_, err = cstore.Export(" " + n1)
	require.NotNil(t, err)
	_, err = cstore.Export(n1 + " ")
	require.NotNil(t, err)
	_, err = cstore.Export("")
	require.NotNil(t, err)
	exported, err := cstore.Export(n1)
	require.Nil(t, err, "%+v", err)

	// import succeeds
	err = cstore.Import(n2, exported)
	require.NoError(t, err)

	// second import fails
	err = cstore.Import(n2, exported)
	require.NotNil(t, err)
}

// TestInMemorySeedPhrase verifies restoring from a seed phrase
func TestInMemorySeedPhrase(t *testing.T) {

	// make the storage with reasonable defaults
	cstore := NewInMemory()

	algo := Secp256k1
	n1, n2 := "lost-key", "found-again"
	p1, p2 := nums, foobar

	// make sure key works with initial password
	info, mnemonic, err := cstore.CreateMnemonic(n1, English, p1, algo)
	require.Nil(t, err, "%+v", err)
	require.Equal(t, n1, info.GetName())
	require.NotEmpty(t, mnemonic)

	// now, let us delete this key
	err = cstore.Delete(n1, p1, false)
	require.Nil(t, err, "%+v", err)
	_, err = cstore.Get(n1)
	require.NotNil(t, err)

	// let us re-create it from the mnemonic-phrase
	params := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	hdPath := params.String()
	newInfo, err := cstore.CreateAccount(n2, mnemonic, DefaultBIP39Passphrase, p2, hdPath, Secp256k1)
	require.NoError(t, err)
	require.Equal(t, n2, newInfo.GetName())
	require.Equal(t, info.GetPubKey().Address(), newInfo.GetPubKey().Address())
	require.Equal(t, info.GetPubKey(), newInfo.GetPubKey())
}

func ExampleNew() {
	// Select the encryption and storage for your cryptostore
	customKeyGenFunc := func(bz []byte, algo SigningAlgo) (tmcrypto.PrivKey, error) {
		var bzArr [32]byte
		copy(bzArr[:], bz)
		return secp256k1.PrivKeySecp256k1(bzArr), nil
	}
	cstore := NewInMemory(WithKeygenFunc(customKeyGenFunc))

	sec := Secp256k1

	// Add keys and see they return in alphabetical order
	bob, _, err := cstore.CreateMnemonic("Bob", English, "friend", sec)
	if err != nil {
		// this should never happen
		fmt.Println(err)
	} else {
		// return info here just like in List
		fmt.Println(bob.GetName())
	}
	_, _, _ = cstore.CreateMnemonic("Alice", English, "secret", sec)
	_, _, _ = cstore.CreateMnemonic("Carl", English, "mitm", sec)
	info, _ := cstore.List()
	for _, i := range info {
		fmt.Println(i.GetName())
	}

	// We need to use passphrase to generate a signature
	tx := []byte("deadbeef")
	sig, pub, err := cstore.Sign("Bob", "friend", tx)
	if err != nil {
		fmt.Println("don't accept real passphrase")
	}

	// and we can validate the signature with publicly available info
	binfo, _ := cstore.Get("Bob")
	if !binfo.GetPubKey().Equals(bob.GetPubKey()) {
		fmt.Println("Get and Create return different keys")
	}

	if pub.Equals(binfo.GetPubKey()) {
		fmt.Println("signed by Bob")
	}
	if !pub.VerifyBytes(tx, sig) {
		fmt.Println("invalid signature")
	}

	// Output:
	// Bob
	// Alice
	// Bob
	// Carl
	// signed by Bob
}

func accAddr(info Info) sdk.AccAddress {
	return (sdk.AccAddress)(info.GetPubKey().Address())
}

var _ tmcrypto.PrivKey = testPriv{}
var _ tmcrypto.PubKey = testPub{}
var testCdc *amino.Codec

type testPriv []byte

func (privkey testPriv) PubKey() tmcrypto.PubKey { return testPub{} }
func (privkey testPriv) Bytes() []byte {
	return testCdc.MustMarshalBinaryBare(privkey)
}
func (privkey testPriv) Sign(msg []byte) ([]byte, error)    { return []byte{}, nil }
func (privkey testPriv) Equals(other tmcrypto.PrivKey) bool { return true }

type testPub []byte

func (key testPub) Address() tmcrypto.Address { return tmcrypto.Address{} }
func (key testPub) Bytes() []byte {
	return testCdc.MustMarshalBinaryBare(key)
}
func (key testPub) VerifyBytes(msg []byte, sig []byte) bool { return true }
func (key testPub) Equals(other tmcrypto.PubKey) bool       { return true }

func TestInMemoryKeygenOverride(t *testing.T) {
	// Save existing codec and reset after test
	cryptoCdc := CryptoCdc
	t.Cleanup(func() {
		CryptoCdc = cryptoCdc
	})

	// Setup testCdc encoding and decoding new key type
	testCdc = codec.New()
	RegisterCodec(testCdc)
	tmamino.RegisterAmino(testCdc)

	// Set up codecs for using new key types
	privName, pubName := "test/priv_name", "test/pub_name"
	tmamino.RegisterKeyType(testPriv{}, privName)
	tmamino.RegisterKeyType(testPub{}, pubName)
	testCdc.RegisterConcrete(testPriv{}, privName, nil)
	testCdc.RegisterConcrete(testPub{}, pubName, nil)
	CryptoCdc = testCdc

	overrideCalled := false
	dummyFunc := func(bz []byte, algo SigningAlgo) (tmcrypto.PrivKey, error) {
		overrideCalled = true
		return testPriv(bz), nil
	}

	kb := NewInMemory(WithKeygenFunc(dummyFunc))

	testName, pw := "name", "testPassword"

	// create new key which will generate with
	info, _, err := kb.CreateMnemonic(testName, English, pw, Secp256k1)
	require.NoError(t, err)
	require.Equal(t, info.GetName(), testName)

	// Assert overridden function was called
	require.True(t, overrideCalled)

	// export private key object
	exported, err := kb.ExportPrivateKeyObject(testName, pw)
	require.Nil(t, err, "%+v", err)

	// require that the key type is the new key
	_, ok := exported.(testPriv)
	require.True(t, ok)

	require.True(t, exported.PubKey().Equals(info.GetPubKey()))
}
