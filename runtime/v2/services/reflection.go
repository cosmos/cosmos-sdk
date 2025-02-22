package services

import (
	"context"

	"google.golang.org/protobuf/types/descriptorpb"

	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
)

// ReflectionService implements the cosmos.reflection.v1 service.
type ReflectionService struct {
	reflectionv1.UnimplementedReflectionServiceServer

	Files *descriptorpb.FileDescriptorSet
}

func (r ReflectionService) FileDescriptors(_ context.Context, _ *reflectionv1.FileDescriptorsRequest) (*reflectionv1.FileDescriptorsResponse, error) {
	return &reflectionv1.FileDescriptorsResponse{
		Files: r.Files.File,
	}, nil
}

var _ reflectionv1.ReflectionServiceServer = &ReflectionService{}
