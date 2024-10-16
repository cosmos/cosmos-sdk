package appconfig

import (
	"strings"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// dynamic resolver allows marshaling gogo proto messages from the gogoproto.HybridResolver as long as those
// files have been imported before calling LoadJSON. There is similar code in autocli, this should probably
// eventually be moved into a library.
type dynamicTypeResolver struct {
	resolver protodesc.Resolver
}

// FindExtensionByName finds an extension type by its full name.
// It first tries to find the extension in the global protobuf registry.
// If not found, it uses the custom resolver to find the descriptor and creates a new extension type.
func (r dynamicTypeResolver) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	ext, err := protoregistry.GlobalTypes.FindExtensionByName(field)
	if err == nil {
		return ext, nil
	}

	desc, err := r.resolver.FindDescriptorByName(field)
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewExtensionType(desc.(protoreflect.ExtensionTypeDescriptor)), nil
}

// FindExtensionByNumber finds an extension type by its message name and field number.
// It first tries to find the extension in the global protobuf registry.
// If not found, it uses the custom resolver to find the message descriptor and searches for the extension.
func (r dynamicTypeResolver) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	ext, err := protoregistry.GlobalTypes.FindExtensionByNumber(message, field)
	if err == nil {
		return ext, nil
	}

	desc, err := r.resolver.FindDescriptorByName(message)
	if err != nil {
		return nil, err
	}

	messageDesc := desc.(protoreflect.MessageDescriptor)
	exts := messageDesc.Extensions()
	n := exts.Len()
	for i := 0; i < n; i++ {
		ext := exts.Get(i)
		if ext.Number() == field {
			return dynamicpb.NewExtensionType(ext), nil
		}
	}

	return nil, protoregistry.NotFound
}

// FindMessageByName finds a message type by its full name.
// It first tries to find the message in the global protobuf registry.
// If not found, it uses the custom resolver to find the descriptor and creates a new message type.
func (r dynamicTypeResolver) FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error) {
	typ, err := protoregistry.GlobalTypes.FindMessageByName(message)
	if err == nil {
		return typ, nil
	}

	desc, err := r.resolver.FindDescriptorByName(message)
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewMessageType(desc.(protoreflect.MessageDescriptor)), nil
}

// FindMessageByURL finds a message type by its URL.
// It extracts the full name from the URL and delegates the search to FindMessageByName.
func (r dynamicTypeResolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		url = url[i+1:]
	}

	return r.FindMessageByName(protoreflect.FullName(url))
}

var (
	_ protoregistry.MessageTypeResolver   = dynamicTypeResolver{}
	_ protoregistry.ExtensionTypeResolver = dynamicTypeResolver{}
)
