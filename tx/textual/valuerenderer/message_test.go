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
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func EmptyCoinMetadataQuerier(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
	return nil, nil
}

func TestMessageValueRenderer_Format(t *testing.T) {
	tr := valuerenderer.NewTextual(EmptyCoinMetadataQuerier)
	mr := valuerenderer.NewMessageValueRenderer(&tr, (&testpb.Foo{}).ProtoReflect().Interface())

	// TODO: add a repeated field to the proto once the renderer for repeated fields is implemented.
	msg := &testpb.Foo{
		FullName: "the scanner",
		ChildBar: &testpb.Bar{
			BarId:      "goku",
			PowerLevel: 9001,
		},
	}

	screens, err := mr.Format(context.Background(), protoreflect.ValueOf((msg).ProtoReflect()))
	assert.NoError(t, err)

	wantScreens := []valuerenderer.Screen{
		{Text: "Foo object"},
		{Text: "Full name: the scanner", Indent: 1},
		{Text: "Child bar: Bar object", Indent: 1},
		{Text: "Bar id: goku", Indent: 2},
		{Text: "Power level: 9'001", Indent: 2},
	}
	assert.Equal(t, wantScreens, screens)
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
			rend := valuerenderer.NewMessageValueRenderer(&tr, (&testpb.Foo{}).ProtoReflect().Interface())

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
