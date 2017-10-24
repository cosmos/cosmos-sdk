package nano

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	crypto "github.com/tendermint/go-crypto"
)

func TestLedgerKeys(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cases := []struct {
		msg, pubkey, sig string
		valid            bool
	}{
		0: {
			msg:    "F00D",
			pubkey: "8E8754F012C2FDB492183D41437FD837CB81D8BBE731924E2E0DAF43FD3F2C93",
			sig:    "787DC03E9E4EE05983E30BAE0DEFB8DB0671DBC2F5874AC93F8D8CA4018F7A42D6F9A9BCEADB422AC8E27CEE9CA205A0B88D22CD686F0A43EB806E8190A3C400",
			valid:  true,
		},
		1: {
			msg:    "DEADBEEF",
			pubkey: "0C45ADC887A5463F668533443C829ED13EA8E2E890C778957DC28DB9D2AD5A6C",
			sig:    "00ED74EED8FDAC7988A14BF6BC222120CBAC249D569AF4C2ADABFC86B792F97DF73C4919BE4B6B0ACB53547273BF29FBF0A9E0992FFAB6CB6C9B09311FC86A00",
			valid:  true,
		},
		2: {
			msg:    "1234567890AA",
			pubkey: "598FC1F0C76363D14D7480736DEEF390D85863360F075792A6975EFA149FD7EA",
			sig:    "59AAB7D7BDC4F936B6415DE672A8B77FA6B8B3451CD95B3A631F31F9A05DAEEE5E7E4F89B64DDEBB5F63DC042CA13B8FCB8185F82AD7FD5636FFDA6B0DC9570B",
			valid:  true,
		},
		3: {
			msg:    "1234432112344321",
			pubkey: "359E0636E780457294CCA5D2D84DB190C3EDBD6879729C10D3963DEA1D5D8120",
			sig:    "616B44EC7A65E7C719C170D669A47DE80C6AC0BB13FBCC89230976F9CC14D4CF9ECF26D4AFBB9FFF625599F1FF6F78EDA15E9F6B6BDCE07CFE9D8C407AC45208",
			valid:  true,
		},
		4: {
			msg:    "12344321123443",
			pubkey: "359E0636E780457294CCA5D2D84DB190C3EDBD6879729C10D3963DEA1D5D8120",
			sig:    "616B44EC7A65E7C719C170D669A47DE80C6AC0BB13FBCC89230976F9CC14D4CF9ECF26D4AFBB9FFF625599F1FF6F78EDA15E9F6B6BDCE07CFE9D8C407AC45208",
			valid:  false,
		},
		5: {
			msg:    "1234432112344321",
			pubkey: "459E0636E780457294CCA5D2D84DB190C3EDBD6879729C10D3963DEA1D5D8120",
			sig:    "616B44EC7A65E7C719C170D669A47DE80C6AC0BB13FBCC89230976F9CC14D4CF9ECF26D4AFBB9FFF625599F1FF6F78EDA15E9F6B6BDCE07CFE9D8C407AC45208",
			valid:  false,
		},
		6: {
			msg:    "1234432112344321",
			pubkey: "359E0636E780457294CCA5D2D84DB190C3EDBD6879729C10D3963DEA1D5D8120",
			sig:    "716B44EC7A65E7C719C170D669A47DE80C6AC0BB13FBCC89230976F9CC14D4CF9ECF26D4AFBB9FFF625599F1FF6F78EDA15E9F6B6BDCE07CFE9D8C407AC45208",
			valid:  false,
		},
	}

	for i, tc := range cases {
		bmsg, err := hex.DecodeString(tc.msg)
		require.NoError(err, "%d", i)

		priv := NewMockKey(tc.msg, tc.pubkey, tc.sig)
		pub := priv.PubKey()
		sig := priv.Sign(bmsg)

		valid := pub.VerifyBytes(bmsg, sig)
		assert.Equal(tc.valid, valid, "%d", i)
	}
}

func TestRealLedger(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	if os.Getenv("WITH_LEDGER") == "" {
		t.Skip("Set WITH_LEDGER to run code on real ledger")
	}
	msg := []byte("kuhehfeohg")

	priv, err := NewPrivKeyLedgerEd25519Ed25519()
	require.Nil(err, "%+v", err)
	pub := priv.PubKey()
	sig := priv.Sign(msg)

	valid := pub.VerifyBytes(msg, sig)
	assert.True(valid)

	// now, let's serialize the key and make sure it still works
	bs := priv.Bytes()
	priv2, err := crypto.PrivKeyFromBytes(bs)
	require.Nil(err, "%+v", err)

	// make sure we get the same pubkey when we load from disk
	pub2 := priv2.PubKey()
	require.Equal(pub, pub2)

	// signing with the loaded key should match the original pubkey
	sig = priv2.Sign(msg)
	valid = pub.VerifyBytes(msg, sig)
	assert.True(valid)

	// make sure pubkeys serialize properly as well
	bs = pub.Bytes()
	bpub, err := crypto.PubKeyFromBytes(bs)
	require.NoError(err)
	assert.Equal(pub, bpub)
}

// TestRealLedgerErrorHandling calls. These tests assume
// the ledger is not plugged in....
func TestRealLedgerErrorHandling(t *testing.T) {
	require := require.New(t)

	if os.Getenv("WITH_LEDGER") != "" {
		t.Skip("Skipping on WITH_LEDGER as it tests unplugged cases")
	}

	// first, try to generate a key, must return an error
	// (no panic)
	_, err := NewPrivKeyLedgerEd25519Ed25519()
	require.Error(err)

	led := PrivKeyLedgerEd25519{} // empty
	// or with some pub key
	ed := crypto.GenPrivKeyEd25519()
	led2 := PrivKeyLedgerEd25519{CachedPubKey: ed.PubKey()}

	// loading these should return errors
	bs := led.Bytes()
	_, err = crypto.PrivKeyFromBytes(bs)
	require.Error(err)

	bs = led2.Bytes()
	_, err = crypto.PrivKeyFromBytes(bs)
	require.Error(err)
}
