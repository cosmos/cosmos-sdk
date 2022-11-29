package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/tx/textual/internal/testpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type repeatedJsonTest struct {
	Proto   *testpb.Qux
	Screens []valuerenderer.Screen
	// TODO Remove once we finished all primitive value renderers parsing
	// https://github.com/cosmos/cosmos-sdk/pull/13696
	// https://github.com/cosmos/cosmos-sdk/pull/13853
	Parses bool
}

func TestRepeatedJsonTestcases(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/repeated.json")
	require.NoError(t, err)

	var testcases []repeatedJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	tr := valuerenderer.NewTextual(mockCoinMetadataQuerier)
	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// Create a context.Context containing all coins metadata, to simulate
			// that they are in state.
			ctx := context.Background()
			rend := valuerenderer.NewMessageValueRenderer(&tr, (&testpb.Qux{}).ProtoReflect().Descriptor())
			require.NoError(t, err)

			screens, err := rend.Format(ctx, protoreflect.ValueOf(tc.Proto.ProtoReflect()))
			require.NoError(t, err)
			require.Equal(t, tc.Screens, screens)

			if tc.Parses {
				val, err := rend.Parse(context.Background(), screens)
				require.NoError(t, err)
				msg := val.Message().Interface()
				require.IsType(t, &testpb.Qux{}, msg)
				baz := msg.(*testpb.Qux)
				require.True(t, proto.Equal(baz, tc.Proto))
			}
		})
	}
}
