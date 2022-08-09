package valuerenderer

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"cosmossdk.io/tx/textual/internal/testpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestMessageValueRenderer_Format(t *testing.T) {
	tr := NewTextual()
	mr := &messageValueRenderer{
		tr: &tr,
	}
	msg := &testpb.Foo{
		FullName: "the scanner",
		ChildBar: &testpb.Bar{
			BarId:      "goku",
			PowerLevel: 9001,
		},
	}

	buf := &bytes.Buffer{}
	assert.NoError(t, mr.Format(context.Background(), protoreflect.ValueOf((msg).ProtoReflect()), buf))

	want := strings.TrimPrefix(`
Foo object
> Full name: the scanner
> Child bar: Bar object
>> Bar id: goku
>> Power level: 9'001
`, "\n")
	assert.Equal(t, want, buf.String())
}
