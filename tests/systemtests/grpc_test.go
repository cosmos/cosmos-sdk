//go:build system_test

package systemtests

import (
	"context"
	"testing"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	systest "cosmossdk.io/systemtests"
)

func TestGRPCReflection(t *testing.T) {
	systest.Sut.ResetChain(t)
	systest.Sut.StartChain(t)

	ctx := context.Background()
	grpcClient, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	descSource := grpcurl.DescriptorSourceFromServer(ctx, grpcreflect.NewClientAuto(ctx, grpcClient))
	services, err := grpcurl.ListServices(descSource)
	require.NoError(t, err)
	require.Greater(t, len(services), 0)
	require.Contains(t, services, "cosmos.staking.v1beta1.Query")
}
