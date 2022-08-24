package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	dur "google.golang.org/protobuf/types/known/durationpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// timestampJsonTest is the type of test cases in the testdata file.
// If the test case has a Proto, try to Format() it. If Error is set, expect
// an error, otherwise match Text, then Parse() the text and expect it to
// match (via proto.Equals()) the original Proto. If the test case has no
// Proto, try to Parse() the Text and expect an error if Error is set.
//
// The Timestamp proto seconds field is int64, but restricted in range
// by convention and will fit within a JSON number.
type timestampJsonTest struct {
	Proto *tspb.Timestamp
	Error bool
	Text  string
}

func TestTimestampJsonTestcases(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/timestamp.json")
	require.NoError(t, err)

	var testcases []timestampJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			rend := valuerenderer.NewTimestampValueRenderer()

			if tc.Proto != nil {
				wr := new(strings.Builder)
				err = rend.Format(context.Background(), protoreflect.ValueOf(tc.Proto.ProtoReflect()), wr)
				if tc.Error {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.Text, wr.String())
			}

			rd := strings.NewReader(tc.Text)
			val, err := rend.Parse(context.Background(), rd)
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

func TestTimestampBadFormat(t *testing.T) {
	rend := valuerenderer.NewTimestampValueRenderer()
	wr := new(strings.Builder)
	err := rend.Format(context.Background(), protoreflect.ValueOf(dur.New(time.Hour).ProtoReflect()), wr)
	require.Error(t, err)
}

type badReader struct{}

var _ io.Reader = badReader{}

func (br badReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("reader error")
}

func TestTimestampBadParse_reader(t *testing.T) {
	rend := valuerenderer.NewTimestampValueRenderer()
	_, err := rend.Parse(context.Background(), badReader{})
	require.ErrorContains(t, err, "reader error")
}
