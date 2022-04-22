package util

import (
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func DescriptorKebabName(descriptor protoreflect.Descriptor) string {
	return strcase.ToKebab(string(descriptor.Name()))
}

func DescriptorDocs(descriptor protoreflect.Descriptor) string {
	return descriptor.ParentFile().SourceLocations().ByDescriptor(descriptor).LeadingComments
}
