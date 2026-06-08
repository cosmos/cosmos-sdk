package services

import (
	"context"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	reflectionpkg "github.com/cosmos/cosmos-sdk/runtime/services/reflection"
)

// ReflectionService implements the cosmos.reflection.v1 service.
type ReflectionService struct {
	reflectionpkg.UnimplementedServiceServer
	files []*descriptorpb.FileDescriptorProto
}

func NewReflectionService() (*ReflectionService, error) {
	fds, err := gogoproto.MergedGlobalFileDescriptors()
	if err != nil {
		return nil, err
	}

	return &ReflectionService{files: fds.File}, nil
}

func (r ReflectionService) FileDescriptors(_ context.Context, _ *reflectionpkg.FileDescriptorsRequest) (*reflectionpkg.FileDescriptorsResponse, error) {
	return &reflectionpkg.FileDescriptorsResponse{
		Files: r.files,
	}, nil
}

var _ reflectionpkg.ServiceServer = &ReflectionService{}
