package gogoreflection

import (
	"bytes"
	"compress/gzip"
	"io"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"golang.org/x/exp/slices"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

var fdFiles *protoregistry.Files

// GetProtodescResolver returns the protodesc.Resolver that is combined by
// merging gogo's file descriptor set and protoregistry's one.
func GetProtodescResolver() protodesc.Resolver {
	// See if there's a cache already
	if fdFiles != nil {
		return fdFiles
	}

	fdSet, err := GetFileDescriptorSet()
	if err != nil {
		panic(err)
	}
	fdFiles, err = protodesc.NewFiles(fdSet)
	if err != nil {
		panic(err)
	}

	return fdFiles
}

// GetFileDescriptorSet returns the global file descriptor set by merging
// the one from gogoproto global registry and from protoregistry.GlobalFiles.
// If there's a name conflict, gogo's descriptor is chosen
func GetFileDescriptorSet() (*descriptorpb.FileDescriptorSet, error) {
	fds := &descriptorpb.FileDescriptorSet{}

	// load gogo proto file descriptors
	allFds := gogoproto.AllFileDescriptors()
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

		// It seems we're registering twice gogo.proto.
		// See Frojdi's comments in server/grpc/gogoreflection/fix_registration.go.
		// Skipping here `gogo.proto` and only including `gogoproto/gogo.proto`.
		if *fd.Name == "gogo.proto" ||
			// If we don't skip this one, we have the error:
			// proto: file "descriptor.proto" has a name conflict over google.protobuf.FileDescriptorSet
			// Is it because we're importing "google.golang.org/protobuf/types/descriptorpb"?
			*fd.Name == "descriptor.proto" {
			continue
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
