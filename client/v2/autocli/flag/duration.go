package flag

import (
	"context"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
)

type durationType struct{}

func (t durationType) NewValue(context.Context, *Builder) Value {
	return &durationValue{}
}

func (t durationType) DefaultValue() string {
	return ""
}

type durationValue struct {
	value *durationpb.Duration
}

func (a durationValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	if a.value == nil {
		return protoreflect.Value{}, nil
	}
	return protoreflect.ValueOfMessage(a.value.ProtoReflect()), nil
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
