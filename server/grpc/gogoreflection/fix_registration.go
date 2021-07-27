package gogoreflection

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"reflect"

	_ "github.com/gogo/protobuf/gogoproto" // required so it does register the gogoproto file descriptor
	gogoproto "github.com/gogo/protobuf/proto"

	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	_ "github.com/regen-network/cosmos-proto" // look above
)

var importsToFix = map[string]string{
	"gogo.proto":   "gogoproto/gogo.proto",
	"cosmos.proto": "cosmos_proto/cosmos.proto",
}

// fixRegistration is required because certain files register themselves in a way
// but are imported by other files in a different way.
// NOTE(fdymylja): This fix should not be needed and should be addressed in some CI.
// Currently every cosmos-sdk proto file is importing gogo.proto as gogoproto/gogo.proto,
// but gogo.proto registers itself as gogo.proto, same goes for cosmos.proto.
func fixRegistration(registeredAs, importedAs string) error {
	raw := gogoproto.FileDescriptor(registeredAs)
	if len(raw) == 0 {
		return fmt.Errorf("file descriptor not found for %s", registeredAs)
	}

	fd, err := decodeFileDesc(raw)
	if err != nil {
		return err
	}

	// fix name
	*fd.Name = importedAs
	fixedRaw, err := compress(fd)
	if err != nil {
		return fmt.Errorf("unable to compress: %w", err)
	}
	gogoproto.RegisterFile(importedAs, fixedRaw)
	return nil
}

func init() {
	// we need to fix the gogoproto filedesc to match the import path
	// in theory this shouldn't be required, generally speaking
	// proto files should be imported as their registration path

	for registeredAs, importedAs := range importsToFix {
		err := fixRegistration(registeredAs, importedAs)
		if err != nil {
			panic(err)
		}
	}
}

// compress compresses the given file descriptor
// nolint: interfacer
func compress(fd *dpb.FileDescriptorProto) ([]byte, error) {
	fdBytes, err := proto.Marshal(fd)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	cw := gzip.NewWriter(buf)
	_, err = cw.Write(fdBytes)
	if err != nil {
		return nil, err
	}
	err = cw.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getFileDescriptor(filePath string) []byte {
	// since we got well known descriptors which are not registered into gogoproto registry
	// but are instead registered into the proto one, we need to check both
	fd := gogoproto.FileDescriptor(filePath)
	if len(fd) != 0 {
		return fd
	}

	return proto.FileDescriptor(filePath)
}

func getMessageType(name string) reflect.Type {
	typ := gogoproto.MessageType(name)
	if typ != nil {
		return typ
	}

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

	protoExts := proto.RegisteredExtensions(m)
	out = make([]int32, 0, len(protoExts))
	for id := range protoExts {
		out = append(out, id)
	}
	return out
}
