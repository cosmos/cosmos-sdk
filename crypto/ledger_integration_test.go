// +build cgo,ledger,test_real_ledger

package crypto

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
	ledger "github.com/zondax/ledger-cosmos-go"
)

const (
	// These tests expect a ledger device initialized to the following mnemonic
	testMnemonic = "equip will roof matter pink blind book anxiety banner elbow sun young"
)

func TestDiscoverDevice(t *testing.T) {
	device, err := discoverLedger()
	require.NoError(t, err)
	require.NotNil(t, device)
	defer device.Close()
}

func TestDiscoverDeviceTwice(t *testing.T) {
	// We expect the second call not to find a device
	device, err := discoverLedger()
	require.NoError(t, err)
	require.NotNil(t, device)
	defer device.Close()

	device2, err := discoverLedger()
	require.Error(t, err)
	require.Equal(t, "no ledger connected", err.Error())
	require.Nil(t, device2)
}

func TestDiscoverDeviceTwiceClosing(t *testing.T) {
	{
		device, err := ledger.FindLedgerCosmosUserApp()
		require.NoError(t, err)
		require.NotNil(t, device)
		require.NoError(t, device.Close())
	}

	device2, err := discoverLedger()
	require.NoError(t, err)
	require.NotNil(t, device2)
	require.NoError(t, device2.Close())
}

func TestPublicKey(t *testing.T) {
	path := *hd.NewFundraiserParams(0, 0)
	priv, err := NewPrivKeyLedgerSecp256k1(path)
	require.Nil(t, err, "%s", err)
	require.NotNil(t, priv)

	pubKeyAddr, err := sdk.Bech32ifyAccPub(priv.PubKey())
	require.NoError(t, err)
	require.Equal(t, "cosmospub1addwnpepqd87l8xhcnrrtzxnkql7k55ph8fr9jarf4hn6udwukfprlalu8lgw0urza0", pubKeyAddr, "Is your device using test mnemonic: %s ?", testMnemonic)
	require.Equal(t, "5075624b6579536563703235366b317b303334464546394344374334433633353838443342303"+
		"3464542353238314239443233324342413334443646334437314145453539323131464642464531464538377d",
		fmt.Sprintf("%x", priv.PubKey()))
}

func TestPublicKeyHDPath(t *testing.T) {
	expectedAnswers := []string{
		"cosmospub1addwnpepqd87l8xhcnrrtzxnkql7k55ph8fr9jarf4hn6udwukfprlalu8lgw0urza0",
		"cosmospub1addwnpepqfsdqjr68h7wjg5wacksmqaypasnra232fkgu5sxdlnlu8j22ztxvlqvd65",
		"cosmospub1addwnpepqw3xwqun6q43vtgw6p4qspq7srvxhcmvq4jrx5j5ma6xy3r7k6dtxmrkh3d",
		"cosmospub1addwnpepqvez9lrp09g8w7gkv42y4yr5p6826cu28ydrhrujv862yf4njmqyyjr4pjs",
		"cosmospub1addwnpepq06hw3enfrtmq8n67teytcmtnrgcr0yntmyt25kdukfjkerdc7lqg32rcz7",
		"cosmospub1addwnpepqg3trf2gd0s2940nckrxherwqhgmm6xd5h4pcnrh4x7y35h6yafmcpk5qns",
		"cosmospub1addwnpepqdm6rjpx6wsref8wjn7ym6ntejet430j4szpngfgc20caz83lu545vuv8hp",
		"cosmospub1addwnpepqvdhtjzy2wf44dm03jxsketxc07vzqwvt3vawqqtljgsr9s7jvydjmt66ew",
		"cosmospub1addwnpepqwystfpyxwcava7v3t7ndps5xzu6s553wxcxzmmnxevlzvwrlqpzz695nw9",
		"cosmospub1addwnpepqw970u6gjqkccg9u3rfj99857wupj2z9fqfzy2w7e5dd7xn7kzzgkgqch0r",
	}

	// Check with device
	for i := uint32(0); i < 10; i++ {
		path := *hd.NewFundraiserParams(0, i)
		fmt.Printf("Checking keys at %v\n", path)

		priv, err := NewPrivKeyLedgerSecp256k1(path)
		require.Nil(t, err, "%s", err)
		require.NotNil(t, priv)

		pubKeyAddr, err := sdk.Bech32ifyAccPub(priv.PubKey())
		require.NoError(t, err)
		require.Equal(t, expectedAnswers[i], pubKeyAddr, "Is your device using test mnemonic: %s ?", testMnemonic)
	}
}

func getFakeTx(accountNumber uint32) []byte {
	tmp := fmt.Sprintf(
		`{"account_number":"%d","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"5000"},"memo":"memo","msgs":[[""]],"sequence":"6"}`,
		accountNumber)

	return []byte(tmp)
}

func TestSignaturesHD(t *testing.T) {
	for account := uint32(0); account < 100; account += 30 {
		msg := getFakeTx(account)

		path := *hd.NewFundraiserParams(account, account/5)
		fmt.Printf("Checking signature at %v\n", path)

		priv, err := NewPrivKeyLedgerSecp256k1(path)
		require.Nil(t, err, "%s", err)

		pub := priv.PubKey()
		sig, err := priv.Sign(msg)
		require.Nil(t, err)

		valid := pub.VerifyBytes(msg, sig)
		require.True(t, valid, "Is your device using test mnemonic: %s ?", testMnemonic)
	}
}

func TestRealLedgerSecp256k1(t *testing.T) {
	msg := getFakeTx(50)
	path := *hd.NewFundraiserParams(0, 0)
	priv, err := NewPrivKeyLedgerSecp256k1(path)
	require.Nil(t, err, "%s", err)

	pub := priv.PubKey()
	sig, err := priv.Sign(msg)
	require.Nil(t, err)

	valid := pub.VerifyBytes(msg, sig)
	require.True(t, valid)

	// now, let's serialize the public key and make sure it still works
	bs := priv.PubKey().Bytes()
	pub2, err := cryptoAmino.PubKeyFromBytes(bs)
	require.Nil(t, err, "%+v", err)

	// make sure we get the same pubkey when we load from disk
	require.Equal(t, pub, pub2)

	// signing with the loaded key should match the original pubkey
	sig, err = priv.Sign(msg)
	require.Nil(t, err)
	valid = pub.VerifyBytes(msg, sig)
	require.True(t, valid)

	// make sure pubkeys serialize properly as well
	bs = pub.Bytes()
	bpub, err := cryptoAmino.PubKeyFromBytes(bs)
	require.NoError(t, err)
	require.Equal(t, pub, bpub)
}
