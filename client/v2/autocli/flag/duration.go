package flag

import (
	"context"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
)

type durationType struct{}

func (durationType) NewValue(context.Context, *Builder) Value {
	return &durationValue{}
}

func (durationType) DefaultValue() string {
	return ""
}

type durationValue struct {
	value *durationpb.Duration
}

func (v durationValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	if v.value == nil {
		return protoreflect.Value{}, nil
	}
	return protoreflect.ValueOfMessage(v.value.ProtoReflect()), nil
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
