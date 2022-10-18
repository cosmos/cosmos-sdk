package autocli

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	"cosmossdk.io/client/v2/internal/testpb"
)

var testCmdDesc = &autocliv1.ServiceCommandDescriptor{
	Service: testpb.Query_ServiceDesc.ServiceName,
	RpcCommandOptions: []*autocliv1.RpcCommandOptions{
		{
			RpcMethod:  "Echo",
			Use:        "echo [pos1] [pos2] [pos3...]",
			Version:    "1.0",
			Alias:      []string{"e"},
			SuggestFor: []string{"eco"},
			Example:    "echo 1 abc {}",
			Short:      "echo echos the value provided by the user",
			Long:       "echo echos the value provided by the user as a proto JSON object with populated with the provided fields and positional arguments",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{
					ProtoField: "positional1",
				},
				{
					ProtoField: "positional2",
				},
				{
					ProtoField: "positional3_varargs",
					Varargs:    true,
				},
			},
			FlagOptions: map[string]*autocliv1.FlagOptions{
				"u32": {
					Name:      "uint32",
					Shorthand: "u",
					Usage:     "some random uint32",
				},
				"i32": {
					Usage:        "some random int32",
					DefaultValue: "3",
				},
				"u64": {
					Usage:             "some random uint64",
					NoOptDefaultValue: "5",
				},
				"deprecated_field": {
					Deprecated: "don't use this",
				},
				"shorthand_deprecated_field": {
					Shorthand:  "s",
					Deprecated: "bad idea",
				},
				"hidden_bool": {
					Hidden: true,
				},
			},
		},
	},
	SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
		// we test the sub-command functionality using the same service with different options
		"deprecatedecho": {
			Service: testpb.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:  "Echo",
					Deprecated: "don't use this",
				},
			},
		},
		"skipecho": {
			Service: testpb.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Echo",
					Skip:      true,
				},
			},
		},
	},
}

func testExec(t *testing.T, args ...string) *testClientConn {
	server := grpc.NewServer()
	testpb.RegisterQueryServer(server, &testEchoServer{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NilError(t, err)
	go func() {
		err := server.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()
	defer server.GracefulStop()
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
	}
	b := &Builder{
		GetClientConn: func(ctx context.Context) grpc.ClientConnInterface {
			return conn
		},
	}
	cmd, err := b.BuildModuleQueryCommand("test", testCmdDesc)
	assert.NilError(t, err)
	cmd.SetArgs(args)
	cmd.SetOut(conn.out)
	assert.NilError(t, cmd.Execute())
	return conn
}

func TestEverything(t *testing.T) {
	conn := testExec(t,
		"echo",
		"1",
		"abc",
		`{"denom":"foo","amount":"1234"}`,
		`{"denom":"bar","amount":"4321"}`,
		"--a-bool",
		"--an-enum", "one",
		"--a-message", `{"bar":"abc", "baz":-3}`,
		"--duration", "4h3s",
		"--uint32", "27",
		"--u-64", "3267246890",
		"--i-32", "-253",
		"--i-64", "-234602347",
		"--str", "def",
		"--timestamp", "2019-01-02T00:01:02Z",
		"--a-coin", `{"denom":"foo","amount":"100000"}`,
		"--an-address", "cosmossdghdsfoi2134sdgh",
		"--bz", "c2RncXdlZndkZ3NkZw==",
		"--page-count-total",
		"--page-key", "MTIzNTQ4N3NnaGRhcw==",
		"--page-limit", "1000",
		"--page-offset", "10",
		"--page-reverse",
		"--bools", "true",
		"--bools", "false,false,true",
		"--enums", "one",
		"--enums", "five",
		"--enums", "two",
		"--strings", "abc",
		"--strings", "xyz",
		"--strings", "xyz,qrs",
		"--durations", "3s",
		"--durations", "5s",
		"--durations", "10h",
		"--some-messages", "{}",
		"--some-messages", `{"bar":"baz"}`,
		"--some-messages", `{"baz":-1}`,
		"--uints", "1,2,3",
		"--uints", "4",
	)
	assert.DeepEqual(t, conn.lastRequest, conn.lastResponse.(*testpb.EchoResponse).Request, protocmp.Transform())
}

func TestOptions(t *testing.T) {
	conn := testExec(t,
		"echo",
		"1", "abc", `{"denom":"foo","amount":"1"}`,
		"-u", "27", // shorthand
		"--u-64", // no opt default value
	)
	lastReq := conn.lastRequest.(*testpb.EchoRequest)
	assert.Equal(t, uint32(27), lastReq.U32) // shorthand got set
	assert.Equal(t, int32(3), lastReq.I32)   // default value got set
	assert.Equal(t, uint64(5), lastReq.U64)  // no opt default value got set
}

func TestHelp(t *testing.T) {
	conn := testExec(t, "-h")
	golden.Assert(t, conn.out.String(), "help-toplevel.golden")

	conn = testExec(t, "echo", "-h")
	golden.Assert(t, conn.out.String(), "help-echo.golden")

	conn = testExec(t, "deprecatedecho", "echo", "-h")
	golden.Assert(t, conn.out.String(), "help-deprecated.golden")

	conn = testExec(t, "skipecho", "-h")
	golden.Assert(t, conn.out.String(), "help-skip.golden")
}

func TestDeprecated(t *testing.T) {
	conn := testExec(t, "echo",
		"1", "abc", `{}`,
		"--deprecated-field", "foo")
	assert.Assert(t, strings.Contains(conn.out.String(), "--deprecated-field has been deprecated"))

	conn = testExec(t, "echo",
		"1", "abc", `{}`,
		"-s", "foo")
	assert.Assert(t, strings.Contains(conn.out.String(), "--shorthand-deprecated-field has been deprecated"))
}

func TestBuildCustomQueryCommand(t *testing.T) {
	b := &Builder{}
	customCommandCalled := false
	cmd, err := b.BuildQueryCommand(map[string]*autocliv1.ModuleOptions{
		"test": {
			Query: testCmdDesc,
		},
	}, map[string]*cobra.Command{
		"test": {Use: "test", Run: func(cmd *cobra.Command, args []string) {
			customCommandCalled = true
		}},
	})
	assert.NilError(t, err)
	cmd.SetArgs([]string{"test", "query"})
	assert.NilError(t, cmd.Execute())
	assert.Assert(t, customCommandCalled)
}

func TestNotFoundErrors(t *testing.T) {
	b := &Builder{}

	// bad service
	_, err := b.BuildModuleQueryCommand("test", &autocliv1.ServiceCommandDescriptor{Service: "foo"})
	assert.ErrorContains(t, err, "can't find service foo")

	// bad method
	_, err = b.BuildModuleQueryCommand("test", &autocliv1.ServiceCommandDescriptor{
		Service:           testpb.Query_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{{RpcMethod: "bar"}},
	})
	assert.ErrorContains(t, err, "rpc method bar not found")

	// bad positional field
	_, err = b.BuildModuleQueryCommand("test", &autocliv1.ServiceCommandDescriptor{
		Service: testpb.Query_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
			{
				RpcMethod: "Echo",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{
					{
						ProtoField: "foo",
					},
				},
			},
		},
	})
	assert.ErrorContains(t, err, "can't find field foo")

	// bad flag field
	_, err = b.BuildModuleQueryCommand("test", &autocliv1.ServiceCommandDescriptor{
		Service: testpb.Query_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
			{
				RpcMethod: "Echo",
				FlagOptions: map[string]*autocliv1.FlagOptions{
					"baz": {},
				},
			},
		},
	})
	assert.ErrorContains(t, err, "can't find field baz")
}

type testClientConn struct {
	*grpc.ClientConn
	t            *testing.T
	lastRequest  interface{}
	lastResponse interface{}
	out          *bytes.Buffer
}

func (t *testClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
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
