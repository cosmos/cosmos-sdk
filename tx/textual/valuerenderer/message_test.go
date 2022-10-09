package valuerenderer

import (
	"context"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/tx/textual/internal/testpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func EmptyCoinMetadataQuerier(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
	return nil, nil
}

func TestMessageValueRenderer_Format(t *testing.T) {
	tr := NewTextual(EmptyCoinMetadataQuerier)

	mr := &messageValueRenderer{
		tr: &tr,
	}
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

	wantScreens := []Screen{
		{Text: "Foo object"},
		{Text: "Full name: the scanner", Indent: 1},
		{Text: "Child bar: Bar object", Indent: 1},
		{Text: "Bar id: goku", Indent: 2},
		{Text: "Power level: 9'001", Indent: 2},
	}
	assert.Equal(t, wantScreens, screens)
}
