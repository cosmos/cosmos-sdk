package nano

import (
	"encoding/hex"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"
)

func parseEdKey(data []byte) (key crypto.PubKey, err error) {
	ed := crypto.PubKeyEd25519{}
	if len(data) < len(ed) {
		return key, errors.Errorf("Key length too short: %d", len(data))
	}
	copy(ed[:], data)
	return ed.Wrap(), nil
}

func parseSig(data []byte) (key crypto.Signature, err error) {
	ed := crypto.SignatureEd25519{}
	if len(data) < len(ed) {
		return key, errors.Errorf("Sig length too short: %d", len(data))
	}
	copy(ed[:], data)
	return ed.Wrap(), nil
}

func TestParseDigest(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cases := []struct {
		output string
		key    string
		sig    string
		valid  bool
	}{
		{
			output: "80028E8754F012C2FDB492183D41437FD837CB81D8BBE731924E2E0DAF43FD3F2C9300CAFE00787DC03E9E4EE05983E30BAE0DEFB8DB0671DBC2F5874AC93F8D8CA4018F7A42D6F9A9BCEADB422AC8E27CEE9CA205A0B88D22CD686F0A43EB806E8190A3C400",
			key:    "8E8754F012C2FDB492183D41437FD837CB81D8BBE731924E2E0DAF43FD3F2C93",
			sig:    "787DC03E9E4EE05983E30BAE0DEFB8DB0671DBC2F5874AC93F8D8CA4018F7A42D6F9A9BCEADB422AC8E27CEE9CA205A0B88D22CD686F0A43EB806E8190A3C400",
			valid:  true,
		},
		{
			output: "800235467890876543525437890796574535467890",
			key:    "",
			sig:    "",
			valid:  false,
		},
	}

	for i, tc := range cases {
		msg, err := hex.DecodeString(tc.output)
		require.Nil(err, "%d: %+v", i, err)

		lKey, lSig, err := parseDigest(msg)
		if !tc.valid {
			assert.NotNil(err, "%d", i)
		} else if assert.Nil(err, "%d: %+v", i, err) {
			key, err := hex.DecodeString(tc.key)
			require.Nil(err, "%d: %+v", i, err)
			sig, err := hex.DecodeString(tc.sig)
			require.Nil(err, "%d: %+v", i, err)

			assert.Equal(key, lKey, "%d", i)
			assert.Equal(sig, lSig, "%d", i)
		}
	}
}

type cryptoCase struct {
	msg   string
	key   string
	sig   string
	valid bool
}

func toBytes(c cryptoCase) (msg, key, sig []byte, err error) {
	msg, err = hex.DecodeString(c.msg)
	if err != nil {
		return
	}
	key, err = hex.DecodeString(c.key)
	if err != nil {
		return
	}
	sig, err = hex.DecodeString(c.sig)
	return
}

func TestCryptoConvert(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cases := []cryptoCase{
		0: {
			msg:   "F00D",
			key:   "8E8754F012C2FDB492183D41437FD837CB81D8BBE731924E2E0DAF43FD3F2C93",
			sig:   "787DC03E9E4EE05983E30BAE0DEFB8DB0671DBC2F5874AC93F8D8CA4018F7A42D6F9A9BCEADB422AC8E27CEE9CA205A0B88D22CD686F0A43EB806E8190A3C400",
			valid: true,
		},
		1: {
			msg:   "DEADBEEF",
			key:   "0C45ADC887A5463F668533443C829ED13EA8E2E890C778957DC28DB9D2AD5A6C",
			sig:   "00ED74EED8FDAC7988A14BF6BC222120CBAC249D569AF4C2ADABFC86B792F97DF73C4919BE4B6B0ACB53547273BF29FBF0A9E0992FFAB6CB6C9B09311FC86A00",
			valid: true,
		},
		2: {
			msg:   "1234567890AA",
			key:   "598FC1F0C76363D14D7480736DEEF390D85863360F075792A6975EFA149FD7EA",
			sig:   "59AAB7D7BDC4F936B6415DE672A8B77FA6B8B3451CD95B3A631F31F9A05DAEEE5E7E4F89B64DDEBB5F63DC042CA13B8FCB8185F82AD7FD5636FFDA6B0DC9570B",
			valid: true,
		},
		3: {
			msg:   "1234432112344321",
			key:   "359E0636E780457294CCA5D2D84DB190C3EDBD6879729C10D3963DEA1D5D8120",
			sig:   "616B44EC7A65E7C719C170D669A47DE80C6AC0BB13FBCC89230976F9CC14D4CF9ECF26D4AFBB9FFF625599F1FF6F78EDA15E9F6B6BDCE07CFE9D8C407AC45208",
			valid: true,
		},
		4: {
			msg:   "12344321123443",
			key:   "359E0636E780457294CCA5D2D84DB190C3EDBD6879729C10D3963DEA1D5D8120",
			sig:   "616B44EC7A65E7C719C170D669A47DE80C6AC0BB13FBCC89230976F9CC14D4CF9ECF26D4AFBB9FFF625599F1FF6F78EDA15E9F6B6BDCE07CFE9D8C407AC45208",
			valid: false,
		},
		5: {
			msg:   "1234432112344321",
			key:   "459E0636E780457294CCA5D2D84DB190C3EDBD6879729C10D3963DEA1D5D8120",
			sig:   "616B44EC7A65E7C719C170D669A47DE80C6AC0BB13FBCC89230976F9CC14D4CF9ECF26D4AFBB9FFF625599F1FF6F78EDA15E9F6B6BDCE07CFE9D8C407AC45208",
			valid: false,
		},
		6: {
			msg:   "1234432112344321",
			key:   "359E0636E780457294CCA5D2D84DB190C3EDBD6879729C10D3963DEA1D5D8120",
			sig:   "716B44EC7A65E7C719C170D669A47DE80C6AC0BB13FBCC89230976F9CC14D4CF9ECF26D4AFBB9FFF625599F1FF6F78EDA15E9F6B6BDCE07CFE9D8C407AC45208",
			valid: false,
		},
	}

	for i, tc := range cases {
		msg, key, sig, err := toBytes(tc)
		require.Nil(err, "%d: %+v", i, err)

		pk, err := parseEdKey(key)
		require.Nil(err, "%d: %+v", i, err)
		psig, err := parseSig(sig)
		require.Nil(err, "%d: %+v", i, err)

		// it is not the signature of the message itself
		valid := pk.VerifyBytes(msg, psig)
		assert.False(valid, "%d", i)

		// but rather of the hash of the msg
		hmsg := hashMsg(msg)
		valid = pk.VerifyBytes(hmsg, psig)
		assert.Equal(tc.valid, valid, "%d", i)
	}
}
