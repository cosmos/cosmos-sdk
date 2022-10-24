package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/tx/signing"
	"cosmossdk.io/tx/textual/valuerenderer"
)

// TODO Remove once we upstream Jim's PR
func EmptyCoinMetadataQuerier(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
	return nil, nil
}

type txJsonTest struct {
	Proto      *txv1beta1.Tx
	SignerData signing.SignerData
	Error      bool
	Text       string
}

func TestTxJsonTestcases(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/tx.json")
	require.NoError(t, err)

	var testcases []txJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tr := valuerenderer.NewTextual(EmptyCoinMetadataQuerier, tc.SignerData)
			rend := valuerenderer.NewTxValueRenderer(&tr)

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
			require.IsType(t, &tspb.Timestamp{}, msg)
			timestamp := msg.(*tspb.Timestamp)
			require.True(t, proto.Equal(timestamp, tc.Proto))
		})
	}
}
