package valuerenderer_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/tx/textual/internal/testpb"
	"cosmossdk.io/tx/textual/valuerenderer"
)

func TestDispatcher(t *testing.T) {
	testcases := []struct {
		name             string
		expErr           bool
		expValueRenderer valuerenderer.ValueRenderer
	}{
		{"UINT32", false, valuerenderer.NewIntValueRenderer()},
		{"UINT64", false, valuerenderer.NewIntValueRenderer()},
		{"SDKINT", false, valuerenderer.NewIntValueRenderer()},
		{"SDKDEC", false, valuerenderer.NewDecValueRenderer()},
		{"BYTES", false, valuerenderer.NewBytesValueRenderer()},
		{"TIMESTAMP", false, valuerenderer.NewTimestampValueRenderer()},
		{"DURATION", false, valuerenderer.NewDurationValueRenderer()},
		{"COIN", false, valuerenderer.NewCoinsValueRenderer(nil)},
		{"COINS", false, valuerenderer.NewCoinsValueRenderer(nil)},
		{"FLOAT", true, nil},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			textual := valuerenderer.NewTextual(nil)
			rend, err := textual.GetValueRenderer(fieldDescriptorFromName(tc.name))

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.IsType(t, tc.expValueRenderer, rend)
			}
		})
	}
}

// fieldDescriptorFromName is like GetADR050ValueRenderer, but taking a Go type
// as input instead of a protoreflect.FieldDescriptor.
func fieldDescriptorFromName(name string) protoreflect.FieldDescriptor {
	a := (&testpb.A{}).ProtoReflect().Descriptor().Fields()
	fd := a.ByName(protoreflect.Name(name))
	if fd == nil {
		panic(fmt.Errorf("no field descriptor for %s", name))
	}

	return fd
}
