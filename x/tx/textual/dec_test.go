package textual_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/x/tx/textual"
)

func TestDecJsonTestcases(t *testing.T) {
	type decimalTest []string
	var testcases []decimalTest
	raw, err := os.ReadFile("./internal/testdata/decimals.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	textual := textual.NewSignModeHandler(nil)

	for _, tc := range testcases {
		tc := tc
		t.Run(tc[0], func(t *testing.T) {
			r, err := textual.GetFieldValueRenderer(fieldDescriptorFromName("SDKDEC"))
			require.NoError(t, err)

			checkNumberTest(t, r, protoreflect.ValueOf(tc[0]), tc[1])
		})
	}
}
