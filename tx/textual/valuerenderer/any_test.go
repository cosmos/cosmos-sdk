package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/tx/textual/valuerenderer"

	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type anyJsonTest struct {
	Proto   map[string]interface{}
	Screens []valuerenderer.Screen
}

func TestAny(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/any.json")
	require.NoError(t, err)

	var testcases []anyJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	tr := valuerenderer.NewTextual(EmptyCoinMetadataQuerier)
	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// Must marshal the proto object back into JSON,
			// then unmarshal with protojson into an Any.
			bz, err := json.Marshal(tc.Proto)
			require.NoError(t, err)
			anyMsg := anypb.Any{}
			err = protojson.Unmarshal(bz, &anyMsg)
			require.NoError(t, err)
			internalMsg, err := anyMsg.UnmarshalNew()
			require.NoError(t, err)

			// Format into screens and check vs expected
			rend := valuerenderer.NewAnyValueRenderer((&tr))
			screens, err := rend.Format(context.Background(), protoreflect.ValueOfMessage(anyMsg.ProtoReflect()))
			require.NoError(t, err)
			require.Equal(t, tc.Screens, screens)

			// Parse back into a google.Protobuf.Any message.
			val, err := rend.Parse(context.Background(), screens)
			require.NoError(t, err)
			parsedMsg := val.Message().Interface()
			require.IsType(t, &anypb.Any{}, parsedMsg)
			parsedAny := parsedMsg.(*anypb.Any)

			// Check for equality of the internal message of the parsed message,
			// to avoid sensitivity to the exact proto encoding bytes.
			parsedInternal, err := parsedAny.UnmarshalNew()
			require.NoError(t, err)
			require.True(t, proto.Equal(internalMsg, parsedInternal))
		})
	}
}
