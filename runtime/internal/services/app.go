package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type AppConfigService struct {
	appv1alpha1.UnimplementedQueryServer
	appConfig *appv1alpha1.Config
	files     *descriptorpb.FileDescriptorSet
}

func NewAppConfigService(appConfig *appv1alpha1.Config) (*AppConfigService, error) {
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

	return &AppConfigService{appConfig: appConfig, files: fds}, nil
}

func (a *AppConfigService) Config(context.Context, *appv1alpha1.QueryConfigRequest) (*appv1alpha1.QueryConfigResponse, error) {
	return &appv1alpha1.QueryConfigResponse{Config: a.appConfig}, nil
}

var _ appv1alpha1.QueryServer = &AppConfigService{}
