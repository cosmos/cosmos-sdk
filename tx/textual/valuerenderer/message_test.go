package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/tx/textual/internal/testpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func EmptyCoinMetadataQuerier(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
	return nil, nil
}

type messageJsonTest struct {
	Proto   *testpb.Foo
	Screens []valuerenderer.Screen
}

func TestMessageJsonTestcases(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/message.json")
	require.NoError(t, err)

	var testcases []messageJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	tr := valuerenderer.NewTextual(EmptyCoinMetadataQuerier)
	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			rend := valuerenderer.NewMessageValueRenderer(&tr, (&testpb.Foo{}).ProtoReflect().Descriptor())

			screens, err := rend.Format(context.Background(), protoreflect.ValueOf(tc.Proto.ProtoReflect()))
			require.NoError(t, err)
			require.Equal(t, tc.Screens, screens)

			val, err := rend.Parse(context.Background(), screens)
			require.NoError(t, err)
			msg := val.Message().Interface()
			require.IsType(t, &testpb.Foo{}, msg)
			foo := msg.(*testpb.Foo)
			require.True(t, proto.Equal(foo, tc.Proto))
		})
	}
}
