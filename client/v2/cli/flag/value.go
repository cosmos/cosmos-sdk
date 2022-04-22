package flag

import (
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type SimpleValue interface {
	pflag.Value
	Get() protoreflect.Value
}

type ListValue interface {
	AppendTo(protoreflect.List)
}
