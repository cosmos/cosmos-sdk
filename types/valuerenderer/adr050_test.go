package valuerenderer_test

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
	"github.com/cosmos/cosmos-sdk/types/valuerenderer/internal/testpb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// formatGoType is like ValueRenderer's Format(), but taking a Go type as input
// value.
func formatGoType(r valuerenderer.ValueRenderer, v interface{}) ([]string, error) {
	switch v.(type) {
	case uint32:
		a := testpb.A{}
		a.ProtoReflect().Set(a.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("UINT32")), protoreflect.ValueOf(v))
		return r.Format(context.Background(), &a)
	case uint64:
		a := testpb.A{}
		a.ProtoReflect().Set(a.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("UINT64")), protoreflect.ValueOf(v))
		return r.Format(context.Background(), &a)
	default:
		return nil, fmt.Errorf("value %s of type %T not recognized", v, v)
	}
}
