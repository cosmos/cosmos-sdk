package nano

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLedgerKeys(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// cryptoCase from sign_test
	cases := []struct {
		msg, pubkey, sig string
		valid            bool
	}{
		{
			msg:    "F00D",
			pubkey: "8E8754F012C2FDB492183D41437FD837CB81D8BBE731924E2E0DAF43FD3F2C93",
			sig:    "787DC03E9E4EE05983E30BAE0DEFB8DB0671DBC2F5874AC93F8D8CA4018F7A42D6F9A9BCEADB422AC8E27CEE9CA205A0B88D22CD686F0A43EB806E8190A3C400",
			valid:  true,
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
