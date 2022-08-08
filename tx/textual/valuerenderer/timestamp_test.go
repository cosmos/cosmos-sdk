package valuerenderer_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	dur "google.golang.org/protobuf/types/known/durationpb"
	ts "google.golang.org/protobuf/types/known/timestamppb"
)

func TestTimestampRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name  string
		proto ts.Timestamp
		text  string
	}{
		{
			name:  "basic_no_frac",
			proto: ts.Timestamp{Seconds: 1136214245},
			text:  "2006-01-02T15:04:05Z",
		},
		{
			name:  "basic_full_frac",
			proto: ts.Timestamp{Seconds: 1136214245, Nanos: 123456789},
			text:  "2006-01-02T15:04:05.123456789Z",
		},
		{
			name:  "basic_trim_frac",
			proto: ts.Timestamp{Seconds: 1136214245, Nanos: 123000000},
			text:  "2006-01-02T15:04:05.123Z",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rend := valuerenderer.NewTimestampValueRenderer()

			wr := new(strings.Builder)
			err := rend.Format(context.Background(), protoreflect.ValueOf(tc.proto.ProtoReflect()), wr)
			require.NoError(t, err)
			require.Equal(t, tc.text, wr.String())

			rd := strings.NewReader(tc.text)
			val, err := rend.Parse(context.Background(), rd)
			require.NoError(t, err)
			msg := val.Message().Interface()
			timestamp, ok := msg.(*ts.Timestamp)
			require.Truef(t, ok, "want Timestamp, got %T", timestamp)
			require.True(t, proto.Equal(timestamp, &tc.proto))
		})
	}
}

func TestTimestampBadFormat(t *testing.T) {
	rend := valuerenderer.NewTimestampValueRenderer()
	wr := new(strings.Builder)
	err := rend.Format(context.Background(), protoreflect.ValueOf(dur.New(time.Hour).ProtoReflect()), wr)
	require.Error(t, err)
}

func TestTimestampBadParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
	}{
		{name: "empty", text: ""},
		{name: "whitespace", text: "   "},
		{name: "garbage", text: "garbage"},
		{name: "silly_americans", text: "11/30/2007"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rend := valuerenderer.NewTimestampValueRenderer()

			rd := strings.NewReader(tc.text)
			_, err := rend.Parse(context.Background(), rd)
			require.Error(t, err)
		})
	}
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
