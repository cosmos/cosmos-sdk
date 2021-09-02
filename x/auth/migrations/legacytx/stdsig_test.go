package legacytx_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

func TestStdSignatureMarshalYAML(t *testing.T) {
	_, pk, _ := testdata.KeyTestPubAddr()
	pkStr := pk.String()

	testCases := []struct {
		sig      legacytx.StdSignature
		expected string
	}{
		{
			legacytx.StdSignature{},
			"|\n  pubkey: \"\"\n  signature: \"\"\n",
		},
		{
			legacytx.StdSignature{PubKey: pk, Signature: []byte("dummySig")},
			fmt.Sprintf("|\n  pubkey: %s\n  signature: 64756D6D79536967\n", pkStr),
		},
		{
			legacytx.StdSignature{PubKey: pk, Signature: nil},
			"pub_key: null\nsignature: null\n",
		},
	}

	for i, tc := range testCases {
		bz, err := yaml.Marshal(tc.sig)
		require.NoError(t, err)
		require.Equal(t, tc.expected, string(bz), "test case #%d", i)
	}
}
