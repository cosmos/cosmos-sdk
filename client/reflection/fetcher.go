package reflection

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
)

// dependencyFetcher is
type dependencyFetcher struct {
	client reflection.ReflectionServiceClient
}

func (c dependencyFetcher) Fetch(filePath string) (desc []byte, err error) {
	resp, err := c.client.ResolveService(context.TODO(), &reflection.ResolveServiceRequest{FileName: filePath})
	if err != nil {
		return nil, err
	}

	return resp.RawDescriptor, nil
}
