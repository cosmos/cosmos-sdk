package textual_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/x/tx/signing/textual"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

type anyJsonTest struct {
	Proto   json.RawMessage
	Screens []textual.Screen
}

func TestAny(t *testing.T) {
	raw, err := os.ReadFile("./internal/testdata/any.json")
	require.NoError(t, err)

	var testcases []anyJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	tr, err := textual.NewSignModeHandler(textual.SignModeOptions{CoinMetadataQuerier: EmptyCoinMetadataQuerier})
	require.NoError(t, err)
	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			anyMsg := anypb.Any{}
			err = protojson.Unmarshal(tc.Proto, &anyMsg)
			require.NoError(t, err)

			// Format into screens and check vs expected
			rend := textual.NewAnyValueRenderer((tr))
			screens, err := rend.Format(context.Background(), protoreflect.ValueOfMessage(anyMsg.ProtoReflect()))
			require.NoError(t, err)
			require.Equal(t, tc.Screens, screens)

			// Parse back into a google.Protobuf.Any message.
			val, err := rend.Parse(context.Background(), screens)
			require.NoError(t, err)
			parsedMsg := val.Message().Interface()
			require.IsType(t, &anypb.Any{}, parsedMsg)
			parsedAny := parsedMsg.(*anypb.Any)
			diff := cmp.Diff(anyMsg, parsedAny, protocmp.Transform())
			require.Empty(t, diff)
		})
	}
}

func TestCustomProtoFiles(t *testing.T) {
	tr, err := textual.NewSignModeHandler(textual.SignModeOptions{
		CoinMetadataQuerier: EmptyCoinMetadataQuerier,
		ProtoTypes:          &protoregistry.Types{}, // Set to empty to force using dynamicpb
	})
	require.NoError(t, err)

	msg := bankv1beta1.MsgSend{FromAddress: "foo"}
	vr, err := tr.GetMessageValueRenderer(msg.ProtoReflect().Descriptor())
	require.NoError(t, err)
	screens, err := vr.Format(context.Background(), protoreflect.ValueOf(msg.ProtoReflect()))
	require.NoError(t, err)

	expScreens := []textual.Screen{
		{Content: "MsgSend object"},
		{Title: "From address", Content: "foo", Indent: 1},
	}
	require.Equal(t, expScreens, screens)
}
