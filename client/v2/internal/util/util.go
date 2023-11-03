package util

import (
	"regexp"
	"runtime/debug"
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

// IsSupportedVersion is used to determine in which version of a module / sdk a rpc was introduced.
// It returns false if the rpc has comment for an higher version than the current one.
func IsSupportedVersion(input string, buildInfo *debug.BuildInfo) bool {
	if input == "" {
		return true // no comment mean it's supported since v0.0.0
	}

	moduleName, version := parseSinceComment(input)
	for _, dep := range buildInfo.Deps {
		if !strings.Contains(dep.Path, moduleName) {
			continue
		}

		if version <= dep.Version {
			return true
		}

		return false
	}

	return true // if cannot find the module consider it's supported
}

// parseSinceComment parses the `// Since: cosmos-sdk v0.xx` comment on rpc.
func parseSinceComment(input string) (string, string) {
	var (
		moduleName string
		version    string
	)

	input = strings.ToLower(input)
	input = strings.ReplaceAll(input, "cosmos sdk", "cosmos-sdk")

	re := regexp.MustCompile(`\/\/\s?since: (\S+) (\S+)`)
	matches := re.FindStringSubmatch(input)
	if len(matches) >= 3 {
		moduleName, version = matches[1], matches[2]

		if !strings.Contains(version, "v") {
			version = "v" + version
		}
	}

	return moduleName, version
}
