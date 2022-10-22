package flag

import (
	"context"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type timestampType struct{}

func (t timestampType) NewValue(context.Context, *Builder) Value {
	return &timestampValue{}
}

func (t timestampType) DefaultValue() string {
	return ""
}

type timestampValue struct {
	value *timestamppb.Timestamp
}

func (t timestampValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	if t.value == nil {
		return protoreflect.Value{}, nil
	}
	return protoreflect.ValueOfMessage(t.value.ProtoReflect()), nil
}

func (v timestampValue) String() string {
	if v.value == nil {
		return ""
	}
	return v.value.AsTime().Format(time.RFC3339)
}

func (v *timestampValue) Set(s string) error {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	v.value = timestamppb.New(t)
	return nil
}

func (v timestampValue) Type() string {
	return "timestamp (RFC 3339)"
}
