package valuerenderer

import (
	"bytes"
	"context"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/tx/textual/internal/testpb"
)

func TestFormatEnum(t *testing.T) {
	enum1 := testpb.Enumeration(1e6 + 17)
	ev1 := protoreflect.ValueOfEnum(enum1.Number())
	er := new(enumRenderer)
	ctx := context.Background()

	buf := new(bytes.Buffer)
	if err := er.Format(ctx, ev1, buf); err != nil {
		t.Fatal(err)
	}
	t.Fatal(buf.String())
}
