package reflection

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
)

// dependencyFetcher is
type dependencyFetcher struct {
	client reflection.ReflectionServiceClient
}

func (c dependencyFetcher) DownloadDescriptorByPath(ctx context.Context, path string) (desc []byte, err error) {
	resp, err := c.client.ResolveService(ctx, &reflection.ResolveServiceRequest{FileName: path})
	if err != nil {
		return nil, err
	}

	return resp.RawDescriptor, nil
}
