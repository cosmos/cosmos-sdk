package ormsql

import (
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	timestampDesc = (&timestamppb.Timestamp{}).ProtoReflect().Descriptor()
	durationDesc  = (&durationpb.Duration{}).ProtoReflect().Descriptor()
)
