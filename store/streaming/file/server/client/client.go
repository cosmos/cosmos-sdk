package client

import (
	"fmt"

	"google.golang.org/grpc"

	pb "github.com/cosmos/cosmos-sdk/store/streaming/file/server/v1beta"
)

// NewClient creates a new gRPC client for the provided endpoint and default configuration
func NewClient(endpoint string) (pb.StateFileClient, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("can not connect with server %v", err)
	}
	return pb.NewStateFileClient(conn), nil
}
