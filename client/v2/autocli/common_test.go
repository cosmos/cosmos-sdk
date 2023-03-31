package autocli

import (
	"bytes"
	"context"
	"net"
	"testing"

	reflectionv2alpha1 "cosmossdk.io/api/cosmos/base/reflection/v2alpha1"
	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/client/v2/internal/testpb"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"gotest.tools/v3/assert"

	"google.golang.org/grpc/credentials/insecure"
)

func testExecCommon(t *testing.T, buildModuleCommand func(string, *Builder) (*cobra.Command, error), args ...string) *testClientConn {
	server := grpc.NewServer()
	testpb.RegisterQueryServer(server, &testEchoServer{})
	reflectionv2alpha1.RegisterReflectionServiceServer(server, &testReflectionServer{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NilError(t, err)
	go func() {
		err := server.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	clientConn, err := grpc.Dial(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NilError(t, err)
	defer func() {
		err := clientConn.Close()
		if err != nil {
			panic(err)
		}
	}()

	conn := &testClientConn{
		ClientConn: clientConn,
		t:          t,
		out:        &bytes.Buffer{},
		errorOut:   &bytes.Buffer{},
	}
	b := &Builder{
		Builder: flag.Builder{
			GetClientConn: func() (grpc.ClientConnInterface, error) {
				return conn, nil
			},
		},
		GetClientConn: func(*cobra.Command) (grpc.ClientConnInterface, error) {
			return conn, nil
		},
		AddQueryConnFlags: flags.AddQueryFlagsToCmd,
		AddTxConnFlags:    flags.AddTxFlagsToCmd,
	}

	cmd, err := buildModuleCommand("test", b)
	assert.NilError(t, err)
	assert.NilError(t, err)
	cmd.SetArgs(args)
	cmd.SetOut(conn.out)
	cmd.SetErr(conn.errorOut)
	cmd.Execute()
	return conn
}

type testReflectionServer struct {
	reflectionv2alpha1.UnimplementedReflectionServiceServer
}

func (t testReflectionServer) GetConfigurationDescriptor(_ context.Context, client *reflectionv2alpha1.GetConfigurationDescriptorRequest) (*reflectionv2alpha1.GetConfigurationDescriptorResponse, error) {
	return &reflectionv2alpha1.GetConfigurationDescriptorResponse{
		Config: &reflectionv2alpha1.ConfigurationDescriptor{
			Bech32AccountAddressPrefix: "cosmos",
		},
	}, nil
}

var _ reflectionv2alpha1.ReflectionServiceServer = testReflectionServer{}
