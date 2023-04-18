package autocli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	"cosmossdk.io/client/v2/internal/testpb"
)

var buildModuleQueryCommand = func(moduleName string, b *Builder) (*cobra.Command, error) {
	cmd := topLevelCmd(moduleName, fmt.Sprintf("Querying commands for the %s module", moduleName))

	err := b.AddQueryServiceCommands(cmd, testCmdDesc)
	return cmd, err
}

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
					Usage:        "some random uint64",
					DefaultValue: "5",
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
				"a_coin": {
					Usage: "some random coin",
				},
				"duration": {
					Usage: "some random duration",
				},
				"bz": {
					Usage: "some bytes",
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

func TestCoin(t *testing.T) {
	conn := testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1",
		"abc",
		"1234foo",
		"4321bar",
		"--a-coin", "100000foo",
		"--duration", "4h3s",
	)
	assert.DeepEqual(t, conn.lastRequest, conn.lastResponse.(*testpb.EchoResponse).Request, protocmp.Transform())
}

func TestEverything(t *testing.T) {
	conn := testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1",
		"abc",
		"123.123123124foo",
		"4321bar",
		"--a-bool",
		"--an-enum", "one",
		"--a-message", `{"bar":"abc", "baz":-3}`,
		"--duration", "4h3s",
		"--uint32", "27",
		"--u64", "3267246890",
		"--i32", "-253",
		"--i64", "-234602347",
		"--str", "def",
		"--timestamp", "2019-01-02T00:01:02Z",
		"--a-coin", "100000foo",
		"--an-address", "cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk",
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
	errOut := conn.errorOut.String()
	res := conn.out.String()
	fmt.Println(errOut, res)
	assert.DeepEqual(t, conn.lastRequest, conn.lastResponse.(*testpb.EchoResponse).Request, protocmp.Transform())
}

func TestJSONParsing(t *testing.T) {
	conn := testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1", "abc", "1foo",
		"--some-messages", `{"bar":"baz"}`,
		"-u", "27", // shorthand
	)
	assert.DeepEqual(t, conn.lastRequest, conn.lastResponse.(*testpb.EchoResponse).Request, protocmp.Transform())

	conn = testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1", "abc", "1foo",
		"--some-messages", "testdata/some_message.json",
		"-u", "27", // shorthand
	)
	assert.DeepEqual(t, conn.lastRequest, conn.lastResponse.(*testpb.EchoResponse).Request, protocmp.Transform())
}

func TestOptions(t *testing.T) {
	conn := testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1", "abc", "123foo",
		"-u", "27", // shorthand
		"--u64", "5", // no opt default value
	)
	lastReq := conn.lastRequest.(*testpb.EchoRequest)
	assert.Equal(t, uint32(27), lastReq.U32) // shorthand got set
	assert.Equal(t, int32(3), lastReq.I32)   // default value got set
	assert.Equal(t, uint64(5), lastReq.U64)  // no opt default value got set
}

func TestBinaryFlag(t *testing.T) {
	// Create a temporary file with some content
	tempFile, err := os.Open("testdata/file.test")
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("this is just a test file")
	if err := tempFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test cases
	tests := []struct {
		name     string
		input    string
		expected []byte
		hasError bool
		err      string
	}{
		{
			name:     "Valid file path with extension",
			input:    tempFile.Name(),
			expected: content,
			hasError: false,
			err:      "",
		},
		{
			name:     "Valid hex-encoded string",
			input:    "68656c6c6f20776f726c64",
			expected: []byte("hello world"),
			hasError: false,
			err:      "",
		},
		{
			name:     "Valid base64-encoded string",
			input:    "SGVsbG8gV29ybGQ=",
			expected: []byte("Hello World"),
			hasError: false,
			err:      "",
		},
		{
			name:     "Invalid input (not a file path or encoded string)",
			input:    "not a file or encoded string",
			expected: nil,
			hasError: true,
			err:      "input string is neither a valid file path, hex, or base64 encoded",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn := testExecCommon(t, buildModuleQueryCommand,
				"echo",
				"1", "abc", `{"denom":"foo","amount":"1"}`,
				"--bz", tc.input,
			)
			errorOut := conn.errorOut.String()
			if errorOut == "" {
				lastReq := conn.lastRequest.(*testpb.EchoRequest)
				assert.DeepEqual(t, tc.expected, lastReq.Bz)
			} else {
				assert.Assert(t, strings.Contains(conn.errorOut.String(), tc.err))
			}
		})
	}
}

func TestAddressValidation(t *testing.T) {
	conn := testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1", "abc", "1foo",
		"--an-address", "cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk",
	)
	assert.Equal(t, "", conn.errorOut.String())

	conn = testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1", "abc", "1foo",
		"--an-address", "regen1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk",
	)
	assert.Assert(t, strings.Contains(conn.errorOut.String(), "Error: invalid argument"))

	conn = testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1", "abc", "1foo",
		"--an-address", "cosmps1BAD_ENCODING",
	)
	assert.Assert(t, strings.Contains(conn.errorOut.String(), "Error: invalid argument"))
}

func TestOutputFormat(t *testing.T) {
	conn := testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1", "abc", "1foo",
		"--output", "json",
	)
	assert.Assert(t, strings.Contains(conn.out.String(), "{"))
	conn = testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1", "abc", "1foo",
		"--output", "text",
	)
	fmt.Println(conn.out.String())
	assert.Assert(t, strings.Contains(conn.out.String(), "  positional1: 1"))
}

func TestHelp(t *testing.T) {
	conn := testExecCommon(t, buildModuleQueryCommand, "-h")
	golden.Assert(t, conn.out.String(), "help-toplevel.golden")

	conn = testExecCommon(t, buildModuleQueryCommand, "echo", "-h")
	golden.Assert(t, conn.out.String(), "help-echo.golden")

	conn = testExecCommon(t, buildModuleQueryCommand, "deprecatedecho", "echo", "-h")
	golden.Assert(t, conn.out.String(), "help-deprecated.golden")

	conn = testExecCommon(t, buildModuleQueryCommand, "skipecho", "-h")
	golden.Assert(t, conn.out.String(), "help-skip.golden")
}

func TestDeprecated(t *testing.T) {
	conn := testExecCommon(t, buildModuleQueryCommand, "echo",
		"1", "abc", `{}`,
		"--deprecated-field", "foo")
	assert.Assert(t, strings.Contains(conn.out.String(), "--deprecated-field has been deprecated"))

	conn = testExecCommon(t, buildModuleQueryCommand, "echo",
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

	buildModuleQueryCommand := func(moduleName string, cmdDescriptor *autocliv1.ServiceCommandDescriptor) (*cobra.Command, error) {
		cmd := topLevelCmd("query", "Querying subcommands")

		err := b.AddMsgServiceCommands(cmd, cmdDescriptor)
		return cmd, err
	}

	// bad service
	_, err := buildModuleQueryCommand("test", &autocliv1.ServiceCommandDescriptor{Service: "foo"})
	assert.ErrorContains(t, err, "can't find service foo")

	// bad method
	_, err = buildModuleQueryCommand("test", &autocliv1.ServiceCommandDescriptor{
		Service:           testpb.Query_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{{RpcMethod: "bar"}},
	})
	assert.ErrorContains(t, err, "rpc method \"bar\" not found")

	// bad positional field
	_, err = buildModuleQueryCommand("test", &autocliv1.ServiceCommandDescriptor{
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
	_, err = buildModuleQueryCommand("test", &autocliv1.ServiceCommandDescriptor{
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
	errorOut     *bytes.Buffer
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
