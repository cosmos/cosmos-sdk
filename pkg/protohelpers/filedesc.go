package protohelpers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"

	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoimpl"
)

func ServiceDescriptorFromGRPCServiceDesc(sd *grpc.ServiceDesc) (protoreflect.ServiceDescriptor, error) {
	fd, err := fileDescriptorFromServiceDesc(sd)
	if err != nil {
		return nil, err
	}
	rsd := fd.Services().ByName(protoreflect.FullName(sd.ServiceName).Name())
	if rsd == nil {
		return nil, fmt.Errorf("service descriptor not found for service: %s", sd.ServiceName)
	}
	return rsd, nil
}

// fileDescriptorFromServiceDesc returns the file descriptor given a gRPC service descriptor
func fileDescriptorFromServiceDesc(sd *grpc.ServiceDesc) (protoreflect.FileDescriptor, error) {
	var compressedFd []byte
	switch meta := sd.Metadata.(type) {
	case string:
		// TODO please remove this once we switch to protov2
		compressedFd = gogoproto.FileDescriptor(meta) // check gogoproto registry
		if len(compressedFd) == 0 {
			compressedFd = proto.FileDescriptor(meta) // check protobuf registry
		}
	case []byte:
		compressedFd = meta
	default:
		return nil, fmt.Errorf("unknown metadata type: %T", meta)
	}

	if len(compressedFd) == 0 {
		return nil, fmt.Errorf("file descriptor not found for %s", sd.ServiceName)
	}
	// decompress file descriptor
	rawFd, err := DecompressFileDescriptor(compressedFd)
	if err != nil {
		return nil, err
	}
	// build fd with a new file and type registry as we don't need to put this into the global registry
	// we just need information
	fd, err := BuildFileDescriptor(rawFd, new(protoregistry.Types), new(protoregistry.Files))
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func BuildFileDescriptor(decompressedFd []byte, types *protoregistry.Types, files *protoregistry.Files) (fd protoreflect.FileDescriptor, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("unable to build descriptor: %v", r)
		}
	}()

	builder := protoimpl.DescBuilder{
		RawDescriptor: decompressedFd,
		TypeResolver:  types,
		FileRegistry:  files,
	}
	fd = builder.Build().File
	return fd, nil
}

func DecompressFileDescriptor(compressed []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("bad gzipped descriptor: %v", err)
	}
	out, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("bad gzipped descriptor: %v", err)
	}
	return out, nil
}
