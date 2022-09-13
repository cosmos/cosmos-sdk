package flag

import (
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Value interface {
	pflag.Value
	HasValue
}

type HasValue interface {
	Get(mutable protoreflect.Value) (protoreflect.Value, error)
}
