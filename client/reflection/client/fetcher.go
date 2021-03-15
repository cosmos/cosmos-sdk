package client

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
)

// protoDownloader implements codec.ProtoImportsDownloader
type protoDownloader struct {
	client reflection.ReflectionServiceClient
}

func (c protoDownloader) DownloadDescriptorByPath(ctx context.Context, path string) (desc []byte, err error) {
	resp, err := c.client.ResolveService(ctx, &reflection.ResolveServiceRequest{FileName: path})
	if err != nil {
		return nil, err
	}

	return resp.RawDescriptor, nil
}
