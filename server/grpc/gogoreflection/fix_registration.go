package gogoreflection

import (
	"reflect"

	_ "github.com/gogo/protobuf/gogoproto" // required so it does register the gogoproto file descriptor
	gogoproto "github.com/gogo/protobuf/proto"

	// nolint: staticcheck
	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	_ "github.com/regen-network/cosmos-proto" // look above
)

func getFileDescriptor(filePath string) []byte {
	// Since we got well known descriptors which are not registered into gogoproto
	// registry but are instead registered into the proto one, we need to check both.
	fd := gogoproto.FileDescriptor(filePath)
	if len(fd) != 0 {
		return fd
	}
	// nolint: staticcheck
	return proto.FileDescriptor(filePath)
}

func getMessageType(name string) reflect.Type {
	typ := gogoproto.MessageType(name)
	if typ != nil {
		return typ
	}
	// nolint: staticcheck
	return proto.MessageType(name)
}

func getExtension(extID int32, m proto.Message) *gogoproto.ExtensionDesc {
	// check first in gogoproto registry
	for id, desc := range gogoproto.RegisteredExtensions(m) {
		if id == extID {
			return desc
		}
	}

	// check into proto registry
	// nolint: staticcheck
	for id, desc := range proto.RegisteredExtensions(m) {
		if id == extID {
			return &gogoproto.ExtensionDesc{
				ExtendedType:  desc.ExtendedType,
				ExtensionType: desc.ExtensionType,
				Field:         desc.Field,
				Name:          desc.Name,
				Tag:           desc.Tag,
				Filename:      desc.Filename,
			}
		}
	}

	return nil
}

func getExtensionsNumbers(m proto.Message) []int32 {
	gogoProtoExts := gogoproto.RegisteredExtensions(m)

	out := make([]int32, 0, len(gogoProtoExts))
	for id := range gogoProtoExts {
		out = append(out, id)
	}
	if len(out) != 0 {
		return out
	}
	// nolint: staticcheck
	protoExts := proto.RegisteredExtensions(m)
	out = make([]int32, 0, len(protoExts))
	for id := range protoExts {
		out = append(out, id)
	}

	return out
}
