package aminojson

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"
)

func TestCosmosInlineJSON(t *testing.T) {
	cases := map[string]struct {
		value      protoreflect.Value
		wantErr    bool
		wantOutput string
	}{
		"supported type - valid JSON object": {
			value:      protoreflect.ValueOfBytes([]byte(`{"test":"value"}`)),
			wantErr:    false,
			wantOutput: `{"test":"value"}`,
		},
		"supported type - valid JSON array": {
			// spaces are normalized away
			value:      protoreflect.ValueOfBytes([]byte(`[1,2,3]`)),
			wantErr:    false,
			wantOutput: `[1,2,3]`,
		},
		"supported type - valid JSON is not normalized": {
			value:      protoreflect.ValueOfBytes([]byte(`[1, 2, 3]`)),
			wantErr:    false,
			wantOutput: `[1, 2, 3]`,
		},
		"supported type - valid JSON array (empty)": {
			value:      protoreflect.ValueOfBytes([]byte(`[]`)),
			wantErr:    false,
			wantOutput: `[]`,
		},
		"supported type - valid JSON number": {
			value:      protoreflect.ValueOfBytes([]byte(`43.72`)),
			wantErr:    false,
			wantOutput: `43.72`,
		},
		"supported type - valid JSON boolean": {
			value:      protoreflect.ValueOfBytes([]byte(`true`)),
			wantErr:    false,
			wantOutput: `true`,
		},
		"supported type - valid JSON null": {
			value:      protoreflect.ValueOfBytes([]byte(`null`)),
			wantErr:    false,
			wantOutput: `null`,
		},
		"supported type - valid JSON string": {
			value:      protoreflect.ValueOfBytes([]byte(`"hey yo"`)),
			wantErr:    false,
			wantOutput: `"hey yo"`,
		},
		"supported type - invalid JSON": {
			value:   protoreflect.ValueOfBytes([]byte(`foo`)),
			wantErr: true,
		},
		"supported type - invalid JSON (empty)": {
			value:   protoreflect.ValueOfBytes([]byte(``)),
			wantErr: true,
		},
		"supported type - invalid JSON (nil bytes)": {
			value:   protoreflect.ValueOfBytes(nil),
			wantErr: true,
		},
		"unsupported type - bool": {
			value:   protoreflect.ValueOfBool(true),
			wantErr: true,
		},
		"unsupported type - int64": {
			value:   protoreflect.ValueOfInt64(1),
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := cosmosInlineJSON(nil, tc.value, &buf)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantOutput, buf.String())
		})
	}
}
