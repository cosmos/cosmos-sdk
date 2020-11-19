//nolint: goconst
package keys

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func init() {
	mintkey.BcryptSecurityParameter = 1
}

const (
	nums   = "1234"
	foobar = "foobar"
)

func TestLanguage(t *testing.T) {
	kb := NewInMemory()
	_, _, err := kb.CreateMnemonic("something", Japanese, "no_pass", Secp256k1)
	assert.Error(t, err)
	assert.Equal(t, "unsupported language: only english is supported", err.Error())
}

func TestCreateAccountInvalidMnemonic(t *testing.T) {
	kb := NewInMemory()
	_, err := kb.CreateAccount(
		"some_account",
		"malarkey pair crucial catch public canyon evil outer stage ten gym tornado",
		"", "", CreateHDPath(0, 0).String(), Secp256k1)
	assert.Error(t, err)
	assert.Equal(t, "Invalid mnemonic", err.Error())
}

func TestCreateLedgerUnsupportedAlgo(t *testing.T) {
	kb := NewInMemory()

	supportedLedgerAlgos := kb.SupportedAlgosLedger()
	for _, supportedAlgo := range supportedLedgerAlgos {
		if Ed25519 == supportedAlgo {
			assert.FailNow(t, "Was not an unsupported algorithm")
		}
	}

	_, err := kb.CreateLedger("some_account", Ed25519, "cosmos", 0, 1)
	assert.Error(t, err)
	assert.Equal(t, "unsupported signing algo", err.Error())
}

func TestCreateLedger(t *testing.T) {
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
	assert.True(t, secpSupported)
	assert.True(t, edSupported)

	ledger, err := kb.CreateLedger("some_account", Secp256k1, "cosmos", 3, 1)

	if err != nil {
		assert.Error(t, err)
		assert.Equal(t, "ledger nano S: support for ledger devices is not available in this executable", err.Error())
		assert.Nil(t, ledger)
		t.Skip("ledger nano S: support for ledger devices is not available in this executable")
		return
	}

	// The mock is available, check that the address is correct
	pubKey := ledger.GetPubKey()
	pk, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubKey)
	assert.NoError(t, err)
	assert.Equal(t, "cosmospub1addwnpepqdszcr95mrqqs8lw099aa9h8h906zmet22pmwe9vquzcgvnm93eqygufdlv", pk)

	// Check that restoring the key gets the same results
	restoredKey, err := kb.Get("some_account")
	assert.NoError(t, err)
	assert.NotNil(t, restoredKey)
	assert.Equal(t, "some_account", restoredKey.GetName())
	assert.Equal(t, TypeLedger, restoredKey.GetType())
	pubKey = restoredKey.GetPubKey()
	pk, err = sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubKey)
	assert.NoError(t, err)
	assert.Equal(t, "cosmospub1addwnpepqdszcr95mrqqs8lw099aa9h8h906zmet22pmwe9vquzcgvnm93eqygufdlv", pk)

	path, err := restoredKey.GetPath()
	assert.NoError(t, err)
	assert.Equal(t, "m/44'/118'/3'/0/1", path.String())
}

// TestKeyManagement makes sure we can manipulate these keys well
func TestKeyManagement(t *testing.T) {
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
	assert.True(t, secpSupported)
	assert.False(t, edSupported)
	assert.True(t, srSupported)

	algo := Secp256k1
	n1, n2, n3 := "personal", "business", "other"
	p1, p2 := nums, "really-secure!@#$"

	// Check empty state
	l, err := cstore.List()
	require.Nil(t, err)
	assert.Empty(t, l)

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

// TestSignVerify does some detailed checks on how we sign and validate
// signatures
func TestSignVerify(t *testing.T) {
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
	require.NoError(t, cstore.ImportPubKey(n3, armor))
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
		key   crypto.PubKey
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
	require.NotNil(t, err)
}

func assertPassword(t *testing.T, cstore Keybase, name, pass, badpass string) {
	getNewpass := func() (string, error) { return pass, nil }
	err := cstore.Update(name, badpass, getNewpass)
	require.NotNil(t, err)
	err = cstore.Update(name, pass, getNewpass)
	require.Nil(t, err, "%+v", err)
}

// TestExportImport tests exporting and importing
func TestExportImport(t *testing.T) {
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

//
func TestExportImportPubKey(t *testing.T) {
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

// TestAdvancedKeyManagement verifies update, import, export functionality
func TestAdvancedKeyManagement(t *testing.T) {
	// make the storage with reasonable defaults
	cstore := NewInMemory()

	algo := Secp256k1
	n1, n2 := "old-name", "new name"
	p1, p2 := nums, foobar

	// make sure key works with initial password
	_, _, err := cstore.CreateMnemonic(n1, English, p1, algo)
	require.Nil(t, err, "%+v", err)
	assertPassword(t, cstore, n1, p1, p2)

	// update password requires the existing password
	getNewpass := func() (string, error) { return p2, nil }
	err = cstore.Update(n1, "jkkgkg", getNewpass)
	require.NotNil(t, err)
	assertPassword(t, cstore, n1, p1, p2)

	// then it changes the password when correct
	err = cstore.Update(n1, p1, getNewpass)
	require.NoError(t, err)
	// p2 is now the proper one!
	assertPassword(t, cstore, n1, p2, p1)

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

// TestSeedPhrase verifies restoring from a seed phrase
func TestSeedPhrase(t *testing.T) {

	// make the storage with reasonable defaults
	cstore := NewInMemory()

	algo := Secp256k1
	n1, n2 := "lost-key", "found-again"
	p1, p2 := nums, foobar

	// make sure key works with initial password
	info, mnemonic, err := cstore.CreateMnemonic(n1, English, p1, algo)
	require.Nil(t, err, "%+v", err)
	require.Equal(t, n1, info.GetName())
	assert.NotEmpty(t, mnemonic)

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

func TestCreateHDPath(t *testing.T) {
	type args struct {
		account uint32
		index   uint32
	}
	tests := []struct {
		name string
		args args
		want hd.BIP44Params
	}{
		{"m/44'/0'/0'/0/0", args{0, 0}, hd.BIP44Params{CoinType: 118, Purpose: 44}},
		{"m/44'/114'/0'/0/0", args{0, 0}, hd.BIP44Params{Purpose: 44, CoinType: 118, Account: 0, AddressIndex: 0}},
		{"m/44'/114'/1'/1/0", args{1, 1}, hd.BIP44Params{Purpose: 44, CoinType: 118, Account: 1, AddressIndex: 1}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			require.Equal(t, tt.want, *CreateHDPath(tt.args.account, tt.args.index))
		})
	}
}

func ExampleNew() {
	// Select the encryption and storage for your cryptostore
	customKeyGenFunc := func(bz []byte, algo SigningAlgo) (crypto.PrivKey, error) {
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
