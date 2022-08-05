package valuerenderer

import (
	"bytes"
	"context"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var intValues = []protoreflect.Value{
	protoreflect.ValueOfString("1000"),
	protoreflect.ValueOfString("99900"),
	protoreflect.ValueOfString("9999999"),
	protoreflect.ValueOfString("999999999999"),
	protoreflect.ValueOfString("9999999999999999999"),
	protoreflect.ValueOfString("100000000000000000000000000000000000000000000000000000000"),
	protoreflect.ValueOfString("77777777777777777777777777777777700"),
	protoreflect.ValueOfString("-77777777777777777777777777777777700"),
	protoreflect.ValueOfString("77777777777777777777777777777777700"),
}

func BenchmarkIntValueRendererFormat(b *testing.B) {
	ctx := context.Background()
	ivr := new(intValueRenderer)
	buf := new(bytes.Buffer)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, value := range intValues {
			if err := ivr.Format(ctx, value, buf); err != nil {
				b.Fatal(err)
			}
		}
		buf.Reset()
	}
}

var decimalValues = []protoreflect.Value{
	protoreflect.ValueOfString("10.00"),
	protoreflect.ValueOfString("999.00"),
	protoreflect.ValueOfString("999.9999"),
	protoreflect.ValueOfString("99999999.9999"),
	protoreflect.ValueOfString("9999999999999999999"),
	protoreflect.ValueOfString("1000000000000000000000000000000000000000000000000000000.00"),
	protoreflect.ValueOfString("77777777777.777777777777777777777700"),
	protoreflect.ValueOfString("-77777777777.777777777777777777777700"),
	protoreflect.ValueOfString("777777777777777777777777.77777777700"),
}

func BenchmarkDecimalValueRendererFormat(b *testing.B) {
	ctx := context.Background()
	dvr := new(decValueRenderer)
	buf := new(bytes.Buffer)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, value := range intValues {
			if err := dvr.Format(ctx, value, buf); err != nil {
				b.Fatal(err)
			}
		}
		buf.Reset()
	}
}

var byteValues = []protoreflect.Value{
	protoreflect.ValueOfBytes(bytes.Repeat([]byte("abc"), 1<<20)),
	protoreflect.ValueOfBytes([]byte("999.00")),
	protoreflect.ValueOfBytes([]byte("999.9999")),
	protoreflect.ValueOfBytes([]byte("99999999.9999")),
	protoreflect.ValueOfBytes([]byte("9999999999999999999")),
	protoreflect.ValueOfBytes([]byte("1000000000000000000000000000000000000000000000000000000.00")),
	protoreflect.ValueOfBytes([]byte("77777777777.777777777777777777777700")),
	protoreflect.ValueOfBytes([]byte("-77777777777.777777777777777777777700")),
	protoreflect.ValueOfBytes([]byte("777777777777777777777777.77777777700")),
}

func BenchmarkBytesValueRendererFormat(b *testing.B) {
	ctx := context.Background()
	bvr := new(bytesValueRenderer)
	buf := new(bytes.Buffer)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, value := range byteValues {
			if err := bvr.Format(ctx, value, buf); err != nil {
				b.Fatal(err)
			}
		}
		buf.Reset()
	}
}
