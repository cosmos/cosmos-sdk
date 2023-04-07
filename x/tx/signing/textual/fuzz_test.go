package textual_test

import (
	"context"
	"testing"

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
