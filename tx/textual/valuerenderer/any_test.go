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
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
)

type anyJsonTest struct {
	Url     string
	Msg     map[string]interface{}
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
			rend := valuerenderer.NewAnyValueRenderer((&tr))

			// Construct the appropriate internal message from its URL
			internalMsgType, err := protoregistry.GlobalTypes.FindMessageByURL(tc.Url)
			require.NoError(t, err)
			internalMsg := internalMsgType.New().Interface()

			// For ease of use, the internal message is shown in the test case as a JSON object,
			// not a string or (shudder) bytes of a proto encoding. So we read that into Go
			// as a map, then render to JSON, then parse as the right message type.
			jsondata, err := json.Marshal(tc.Msg)
			require.NoError(t, err)
			err = json.Unmarshal(jsondata, &internalMsg)
			require.NoError(t, err)

			// Now embed the internal message in a google.protobuf.Any message.
			anyMsg, err := anypb.New(internalMsg)
			require.NoError(t, err)

			// Format into screens and check vs expected
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
			require.Equal(t, tc.Url, parsedAny.TypeUrl)
			parsedInternal, err := parsedAny.UnmarshalNew()
			require.NoError(t, err)
			require.True(t, proto.Equal(internalMsg, parsedInternal))
		})
	}
}
