//go:build system_test

package systemtests

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/fullstorydev/grpcurl" //nolint:staticcheck: input in grpcurl
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	systest "cosmossdk.io/systemtests"
)

func TestGRPC(t *testing.T) {
	systest.Sut.ResetChain(t)
	systest.Sut.StartChain(t)

	ctx := context.Background()
	grpcClient, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer grpcClient.Close()

	// test grpc reflection
	descSource := grpcurl.DescriptorSourceFromServer(ctx, grpcreflect.NewClientAuto(ctx, grpcClient))
	services, err := grpcurl.ListServices(descSource)
	require.NoError(t, err)
	require.Greater(t, len(services), 0)
	require.Contains(t, services, "cosmos.staking.v1beta1.Query")

	// test query invocation
	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.FormatText, descSource, os.Stdin, grpcurl.FormatOptions{})
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	h := &grpcurl.DefaultEventHandler{
		Out:            buf,
		Formatter:      formatter,
		VerbosityLevel: 0,
	}

	err = grpcurl.InvokeRPC(ctx, descSource, grpcClient, "cosmos.staking.v1beta1.Query/Params", nil, h, rf.Next)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "max_validators")
}

func TestGRPCQueryAutoCLIOptions(t *testing.T) {
	systest.Sut.ResetChain(t)
	systest.Sut.StartChain(t)

	ctx := context.Background()
	grpcClient, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	descSource := grpcurl.DescriptorSourceFromServer(ctx, grpcreflect.NewClientAuto(ctx, grpcClient))

	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.FormatText, descSource, os.Stdin, grpcurl.FormatOptions{})
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	h := &grpcurl.DefaultEventHandler{
		Out:            buf,
		Formatter:      formatter,
		VerbosityLevel: 0,
	}

	err = grpcurl.InvokeRPC(ctx, descSource, grpcClient, "cosmos.autocli.v1.Query/AppOptions", nil, h, rf.Next)
	require.NoError(t, err)
}
