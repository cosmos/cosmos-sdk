package codec

import (
	"context"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
)

// DependencyFetcher represents an interface
// that fetches missing proto imports
type DependencyFetcher interface {
	// Fetch gathers the missing dependency raw bytes
	Fetch(filePath string) (desc []byte, err error)
}

func newCosmosFetcher(grpcEndpoint string) (DependencyFetcher, error) {
	conn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return cosmosFetcher{client: reflection.NewReflectionServiceClient(conn)}, nil
}

type cosmosFetcher struct {
	client reflection.ReflectionServiceClient
}

func (c cosmosFetcher) Fetch(filePath string) (desc []byte, err error) {
	resp, err := c.client.ResolveService(context.TODO(), &reflection.ResolveServiceRequest{FileName: filePath})
	if err != nil {
		return nil, err
	}

	return resp.RawDescriptor, nil
}
