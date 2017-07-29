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
			output: "800204338EB1DD3CCDEE1F6FB586F66E640F56FFDD14537A3F0ED9EEEDF10B528FE4195FD17AC9EDAE9718A50196A1459E2434C1E53F1238F4CFDF177FAFBA8B39249B00CAFE00FFDEA42A699205B217004E7E2FFB884E174A548D644116F4B20469CBC32F60A9CB0EEB5BB6A7F266BD0F6A0A99A45B4F18F0F477AED7C854C404EF43530DAB00",
			key:    "04338EB1DD3CCDEE1F6FB586F66E640F56FFDD14537A3F0ED9EEEDF10B528FE4195FD17AC9EDAE9718A50196A1459E2434C1E53F1238F4CFDF177FAFBA8B39249B",
			sig:    "FFDEA42A699205B217004E7E2FFB884E174A548D644116F4B20469CBC32F60A9CB0EEB5BB6A7F266BD0F6A0A99A45B4F18F0F477AED7C854C404EF43530DAB00",
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
			msg:   "00",
			key:   "04338EB1DD3CCDEE1F6FB586F66E640F56FFDD14537A3F0ED9EEEDF10B528FE4195FD17AC9EDAE9718A50196A1459E2434C1E53F1238F4CFDF177FAFBA8B39249B",
			sig:   "FFDEA42A699205B217004E7E2FFB884E174A548D644116F4B20469CBC32F60A9CB0EEB5BB6A7F266BD0F6A0A99A45B4F18F0F477AED7C854C404EF43530DAB00",
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

		// how do i make this valid?
		valid := pk.VerifyBytes(msg, psig)
		assert.Equal(tc.valid, valid, "%d", i)
	}
}
