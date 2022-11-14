package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/tx/textual/internal/testpb"
	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type enumTest struct {
	Proto json.RawMessage
	Text  string
}

func TestEnumJsonTestcases(t *testing.T) {
	var testcases []enumTest
	raw, err := os.ReadFile("../internal/testdata/enum.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	textual := valuerenderer.NewTextual(nil)

	for _, tc := range testcases {
		t.Run(tc.Text, func(t *testing.T) {
			m := &testpb.Baz{}
			err := protojson.Unmarshal(tc.Proto, m)
			require.NoError(t, err)

			fd := getFd(m)
			valrend, err := textual.GetValueRenderer(fd)
			require.NoError(t, err)

			screens, err := valrend.Format(context.Background(), m.ProtoReflect().Get(fd))
			require.NoError(t, err)
			require.Equal(t, 1, len(screens))
			require.Equal(t, tc.Text, screens[0].Text)

			// Round trip
			// val, err := valrend.Parse(context.Background(), screens)
			// require.NoError(t, err)
			// require.Equal(t, tc.base64, base64.StdEncoding.EncodeToString(val.Bytes()))
		})
	}
}

// getFd returns the field descriptor on Baz whose value is set
func getFd(m *testpb.Baz) protoreflect.FieldDescriptor {
	for i := 1; i <= 3; i++ {
		fd := m.ProtoReflect().Descriptor().Fields().ByNumber(protowire.Number(i))
		value := m.ProtoReflect().Get(fd).Enum()
		if value > 0 {
			return fd
		}
	}

	panic(fmt.Errorf("no enums set on %+v", m))
}
