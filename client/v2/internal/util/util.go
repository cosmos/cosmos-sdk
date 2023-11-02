package util

import (
	"regexp"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"

	"cosmossdk.io/client/v2/internal/strcase"
)

// DescriptorName returns the name of the descriptor in kebab case.
func DescriptorKebabName(descriptor protoreflect.Descriptor) string {
	return strcase.ToKebab(string(descriptor.Name()))
}

// DescriptorDocs returns the leading comments of the descriptor.
// TODO this does not work, to fix.
func DescriptorDocs(descriptor protoreflect.Descriptor) string {
	return descriptor.ParentFile().SourceLocations().ByDescriptor(descriptor).LeadingComments
}

func ResolveMessageType(resolver protoregistry.MessageTypeResolver, descriptor protoreflect.MessageDescriptor) protoreflect.MessageType {
	typ, err := resolver.FindMessageByName(descriptor.FullName())
	if err == nil {
		return typ
	}

	return dynamicpb.NewMessageType(descriptor)
}

// ParseSinceComment parses the `// Since: cosmos-sdk v0.xx` comment on rpc.
// It is used to determine in which version of a module / sdk a rpc was introduced.
func ParseSinceComment(input string) (string, string) {
	var (
		moduleName string
		version    string
	)

	re := regexp.MustCompile(`\/\/ since: (\S+) (\S+)`)
	matches := re.FindStringSubmatch(strings.ToLower(input))
	if len(matches) >= 3 {
		moduleName = matches[1]
		version = strings.TrimLeft(matches[2], "v")
	}

	return moduleName, version
}
