package aminojson

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"
)

func TestCosmosBytesAsString(t *testing.T) {
	cases := map[string]struct {
		value      protoreflect.Value
		wantErr    bool
		wantOutput string
	}{
		"valid bytes - json": {
			value:      protoreflect.ValueOfBytes([]byte(`{"test":"value"}`)),
			wantErr:    false,
			wantOutput: `{"test":"value"}`,
		},
		"valid bytes - string": {
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
			err := cosmosBytesAsString(nil, tc.value, &buf)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantOutput, buf.String())
		})
	}
}
