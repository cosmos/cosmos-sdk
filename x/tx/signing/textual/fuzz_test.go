package textual_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	"cosmossdk.io/x/tx/signing/textual"
)

func FuzzIntValueRendererParse(f *testing.F) {
	if testing.Short() {
		f.Skip()
	}

	// 1. Firstly add some seeds
	f.Add("10.11")
	f.Add("-10.11")
	f.Add("0.999999")
	f.Add(".999999")
	f.Add("1'000.999999")
	f.Add("1'000'111")
	f.Add("340'282'366'920'938'463'463'374'607'431'768'211'455")

	// 2. Next setup and run the fuzzer.
	ivr := textual.NewIntValueRenderer(fieldDescriptorFromName("UINT64"))
	ctx := context.Background()
	f.Fuzz(func(t *testing.T, input string) {
		_, _ = ivr.Parse(ctx, []textual.Screen{{Content: input}})
	})
}

func FuzzTimestampValueRendererParse(f *testing.F) {
	if testing.Short() {
		f.Skip()
	}

	// 1. Firstly add some seed valid content.
	f.Add("2006-01-02T15:04:05Z")
	f.Add("1970-01-01T00:00:00.00000001Z")
	f.Add("2022-07-14T11:22:20.983Z")
	f.Add("1969-12-31T23:59:59Z")

	// 2. Now fuzz it.
	tvr := textual.NewTimestampValueRenderer()
	ctx := context.Background()
	f.Fuzz(func(t *testing.T, input string) {
		_, _ = tvr.Parse(ctx, []textual.Screen{{Content: input}})
	})
}

func FuzzTimestampJSONParseToParseRoundTrip(f *testing.F) {
	// 1. Use the seeds from testdata and mutate them.
	seed, err := os.ReadFile("./internal/testdata/timestamp.json")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)

	f.Fuzz(func(t *testing.T, input []byte) {
		var testCases []timestampJSONTest
		if err := json.Unmarshal(input, &testCases); err != nil {
			return
		}

		for _, tc := range testCases {
			rend := textual.NewTimestampValueRenderer()

			// If it successfully JSON unmarshals let's test it out.
			var screens []textual.Screen
			var err error

			if tc.Proto != nil {
				screens, err = rend.Format(context.Background(), protoreflect.ValueOf(tc.Proto.ProtoReflect()))
				if err != nil {
					continue
				}
			}

			val, err := rend.Parse(context.Background(), screens)
			if err != nil {
				continue
			}

			msg := val.Message().Interface()
			gotTs, ok := msg.(*tspb.Timestamp)
			if !ok {
				t.Fatalf("Wrong type for timestamp: %T", msg)
			}
			// Please avoid using proto.Equal to compare timestamps given they aren't
			// in standardized form and will produce false positives for example given input:
			//  []byte(`[{"proto":{"nanos":1000000000}}]`)
			// Per issue: https://github.com/cosmos/cosmos-sdk/issues/15761
			if !gotTs.AsTime().Equal(tc.Proto.AsTime()) {
				t.Fatalf("Roundtrip mismatch\n\tGot:  %#v\n\tWant: %#v", gotTs, tc.Proto)
			}
		}
	})
}
