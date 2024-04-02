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
		// This test case is a bit tricky. Conceptually it makes no sense for this
		// to pass. The question is just where the JSON validity check is done.
		// In wasmd we have it in the message field validation. But it might make
		// sense to move it here instead.
		// For now it's better to consider this case undefined behaviour.
		"supported type - invalid JSON": {
			value:      protoreflect.ValueOfBytes([]byte(`foo`)),
			wantErr:    false,
			wantOutput: `foo`,
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
