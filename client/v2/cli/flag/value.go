package flag

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type SimpleValue interface {
	Get() protoreflect.Value
}

type ListValue interface {
	AppendTo(protoreflect.List)
}
