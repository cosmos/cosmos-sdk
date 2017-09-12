package nano

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		{
			msg:   "F00D",
			key:   "8E8754F012C2FDB492183D41437FD837CB81D8BBE731924E2E0DAF43FD3F2C93",
			sig:   "787DC03E9E4EE05983E30BAE0DEFB8DB0671DBC2F5874AC93F8D8CA4018F7A42D6F9A9BCEADB422AC8E27CEE9CA205A0B88D22CD686F0A43EB806E8190A3C400",
			valid: true,
		},
	}

	for i, tc := range cases {
		msg, key, sig, err := toBytes(tc)
		require.Nil(err, "%d: %+v", i, err)

		pk, err := parseKey(key)
		require.Nil(err, "%d: %+v", i, err)
		psig, err := parseSig(sig)
		require.Nil(err, "%d: %+v", i, err)

		// it is not the signature of the message itself
		valid := pk.VerifyBytes(msg, psig)
		assert.NotEqual(tc.valid, valid, "%d", i)

		// but rather of the hash of the msg
		hmsg := hashMsg(msg)
		valid = pk.VerifyBytes(hmsg, psig)
		assert.Equal(tc.valid, valid, "%d", i)
	}
}
