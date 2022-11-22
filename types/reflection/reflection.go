package reflection

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/cosmos/gogoproto/proto"
	"golang.org/x/exp/slices"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// GetFileDescriptorSet returns the global file descriptor set by merging
// the one from gogoproto global registry and from protoregistry.GlobalFiles.
func GetFileDescriptorSet() (*descriptorpb.FileDescriptorSet, error) {
	fmt.Println("Calling GetFileDescriptorSet")
	fds := &descriptorpb.FileDescriptorSet{}

	// load gogo proto file descriptors
	allFds := proto.AllFileDescriptors()
	haveFileDescriptor := map[string]bool{}
	for _, compressedBz := range allFds {
		rdr, err := gzip.NewReader(bytes.NewReader(compressedBz))
		if err != nil {
			return nil, err
		}

		bz, err := io.ReadAll(rdr)
		if err != nil {
			return nil, err
		}

		fd := &descriptorpb.FileDescriptorProto{}
		err = protov2.Unmarshal(bz, fd)
		if err != nil {
			return nil, err
		}

		fds.File = append(fds.File, fd)
		haveFileDescriptor[*fd.Name] = true
	}

	// load any protoregistry file descriptors not in gogo
	protoregistry.GlobalFiles.RangeFiles(func(fileDescriptor protoreflect.FileDescriptor) bool {
		if !haveFileDescriptor[fileDescriptor.Path()] {
			fds.File = append(fds.File, protodesc.ToFileDescriptorProto(fileDescriptor))
		}
		return true
	})

	slices.SortFunc(fds.File, func(x, y *descriptorpb.FileDescriptorProto) bool {
		return *x.Name < *y.Name
	})

	return fds, nil
}
