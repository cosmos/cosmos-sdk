//go:build gofuzz || go1.18

package tests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/require"
)

func FuzzUnknownProto(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		msg := new(testdata.TestVersion2)
		resolver := new(unknownproto.DefaultAnyResolver)
		_, err := unknownproto.RejectUnknownFields(b, msg, true, resolver)
		require.NoError(t, err)

		_, err = unknownproto.RejectUnknownFields(b, msg, false, resolver)
		require.NoError(t, err)
	})
}
