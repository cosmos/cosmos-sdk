package decode_test

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/tx/internal/testpb"
)

func TestDecode(t *testing.T) {
	msg := &testpb.A{
		UINT32:    0,
		UINT64:    0,
		INT32:     0,
		INT64:     0,
		SDKINT:    "",
		SDKDEC:    "",
		COIN:      nil,
		COINS:     nil,
		BYTES:     nil,
		TIMESTAMP: nil,
		DURATION:  nil,
		ENUM:      0,
		ANY:       nil,
		SINT32:    0,
		SINT64:    0,
		SFIXED32:  0,
		FIXED32:   0,
		FLOAT:     0,
		SFIXED64:  0,
		FIXED64:   0,
		DOUBLE:    0,
		MAP:       nil,
	}
	_, err := proto.Marshal(msg)
	require.NoError(t, err)
	// TODO implement tests once a tx builder is available
}
