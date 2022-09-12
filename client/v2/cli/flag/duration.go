package flag

import (
	"context"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
)

var durationType = Type{
	NewValue: func(ctx context.Context, builder *Builder) Value {
		return &durationValue{}
	},
	DefaultValue: "",
}

type durationValue struct {
	value *durationpb.Duration
}

func (t durationValue) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfMessage(t.value.ProtoReflect()))
}

func (t durationValue) Get() (protoreflect.Value, error) {
	if t.value == nil {
		return protoreflect.Value{}, nil
	}
	return protoreflect.ValueOfMessage(t.value.ProtoReflect()), nil
}

func (v durationValue) String() string {
	if v.value == nil {
		return ""
	}
	return v.value.AsDuration().String()
}

func (v *durationValue) Set(s string) error {
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	v.value = durationpb.New(dur)
	return nil
}

func (v durationValue) Type() string {
	return "duration"
}
