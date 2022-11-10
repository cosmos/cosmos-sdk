package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	dpb "google.golang.org/protobuf/types/known/durationpb"
)

type durationTest struct {
	Proto *dpb.Duration
	Text  string
	Error bool
}

func TestDurationJSON(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/duration.json")
	require.NoError(t, err)

	var testcases []durationTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			rend := valuerenderer.NewDurationValueRenderer()

			var screens []valuerenderer.Screen
			if tc.Proto != nil {
				screens, err = rend.Format(context.Background(), protoreflect.ValueOf(tc.Proto.ProtoReflect()))
				if tc.Error {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, 1, len(screens))
				require.Equal(t, tc.Text, screens[0].Text)
			}

			val, err := rend.Parse(context.Background(), screens)
			if tc.Error {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			msg := val.Message().Interface()
			require.IsType(t, &dpb.Duration{}, msg)
			duration := msg.(*dpb.Duration)
			require.True(t, proto.Equal(duration, tc.Proto), "%v vs %v", duration, tc.Proto)
		})
	}
}
