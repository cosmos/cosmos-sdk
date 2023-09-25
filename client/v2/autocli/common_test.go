package autocli

import (
	"bytes"
	"context"
	"net"
	"testing"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gotest.tools/v3/assert"

	reflectionv2alpha1 "cosmossdk.io/api/cosmos/base/reflection/v2alpha1"
	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/client/v2/internal/testpb"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type fixture struct {
	conn *testClientConn
	b    *Builder
}

func initFixture(t *testing.T) *fixture {
	t.Helper()
	home := t.TempDir()
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

	appCodec := moduletestutil.MakeTestEncodingConfig().Codec
	kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, home, nil, appCodec)
	assert.NilError(t, err)

	var initClientCtx client.Context
	initClientCtx = initClientCtx.
		WithAddressCodec(addresscodec.NewBech32Codec("cosmos")).
		WithValidatorAddressCodec(addresscodec.NewBech32Codec("cosmosvaloper")).
		WithConsensusAddressCodec(addresscodec.NewBech32Codec("cosmosvalcons")).
		WithKeyring(kr).
		WithKeyringDir(home).
		WithHomeDir(home).
		WithViper("")

	conn := &testClientConn{ClientConn: clientConn}
	b := &Builder{
		Builder: flag.Builder{
			TypeResolver: protoregistry.GlobalTypes,
			FileResolver: protoregistry.GlobalFiles,
			ClientCtx:    &initClientCtx,
			Keyring:      kr,
		},
		GetClientConn: func(*cobra.Command) (grpc.ClientConnInterface, error) {
			return conn, nil
		},
		AddQueryConnFlags: flags.AddQueryFlagsToCmd,
		AddTxConnFlags:    flags.AddTxFlagsToCmd,
	}
	assert.NilError(t, b.ValidateAndComplete())

	return &fixture{
		conn: conn,
		b:    b,
	}
}

func runCmd(conn *testClientConn, b *Builder, command func(moduleName string, b *Builder) (*cobra.Command, error), args ...string) (*bytes.Buffer, error) {
	out := &bytes.Buffer{}
	cmd, err := command("test", b)
	if err != nil {
		return out, err
	}

	cmd.SetArgs(args)
	cmd.SetOut(out)
	return out, cmd.Execute()
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

type testClientConn struct {
	*grpc.ClientConn
	lastRequest  interface{}
	lastResponse interface{}
}

func (t *testClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	err := t.ClientConn.Invoke(ctx, method, args, reply, opts...)
	t.lastRequest = args
	t.lastResponse = reply
	return err
}

type testEchoServer struct {
	testpb.UnimplementedQueryServer
}

func (t testEchoServer) Echo(_ context.Context, request *testpb.EchoRequest) (*testpb.EchoResponse, error) {
	return &testpb.EchoResponse{Request: request}, nil
}

var _ testpb.QueryServer = testEchoServer{}
