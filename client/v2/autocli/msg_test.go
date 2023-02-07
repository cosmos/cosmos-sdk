package autocli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gotest.tools/v3/assert"

	"cosmossdk.io/client/v2/internal/testpb"
)

var testCmdMsgDesc = &autocliv1.ServiceCommandDescriptor{
	Service: testpb.Msg_ServiceDesc.ServiceName,
	RpcCommandOptions: []*autocliv1.RpcCommandOptions{
		{
			RpcMethod:  "Send",
			Use:        "send [pos1] [pos2] [pos3...]",
			Version:    "1.0",
			Alias:      []string{"s"},
			SuggestFor: []string{"send"},
			Example:    "send 1 abc {}",
			Short:      "send msg the value provided by the user",
			Long:       "send msg the value provided by the user as a proto JSON object with populated with the provided fields and positional arguments",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{
					ProtoField: "positional1",
				},
				{
					ProtoField: "positional2",
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
					Shorthand:  "d",
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
		"deprecatedmsg": {
			Service: testpb.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:  "Send",
					Deprecated: "dont use this",
				},
			},
		},
	},
}

func testMsgExec(t *testing.T, args ...string) *testClientConn {
	server := grpc.NewServer()
	testpb.RegisterMsgServer(server, &testMessageServer{})
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
		GetClientConn: func(*cobra.Command) (grpc.ClientConnInterface, error) {
			return conn, nil
		},
	}
	cmd, err := b.BuildModuleMsgCommand("test", testCmdMsgDesc)
	assert.NilError(t, err)
	cmd.SetArgs(args)
	cmd.SetOut(conn.out)
	assert.NilError(t, cmd.Execute())
	return conn
}

func TestMsgOptions(t *testing.T) {
	conn := testMsgExec(t,
		"send", "5", "6",
		"--uint32", "7",
		"--u64", "5",
	)
	response := conn.out.String()
	var output testpb.MsgRequest
	json.Unmarshal([]byte(response), &output)
	assert.Equal(t, output.GetU32(), uint32(7))
	assert.Equal(t, output.GetU64(), uint64(0))
	assert.Equal(t, output.GetPositional1(), int32(5))
	assert.Equal(t, output.GetPositional2(), "6")
}

func TestDeprecatedMsg(t *testing.T) {
	conn := testMsgExec(t, "send",
		"1", "abc",
		"--deprecated-field", "foo")
	assert.Assert(t, strings.Contains(conn.out.String(), "--deprecated-field has been deprecated"))

	conn = testMsgExec(t, "send",
		"1", "abc",
		"-d", "foo")
	assert.Assert(t, strings.Contains(conn.out.String(), "--shorthand-deprecated-field has been deprecated"))
}

func TestEverythingMsg(t *testing.T) {
	conn := testMsgExec(t,
		"send",
		"1",
		//"abc",
		//`{"denom":"foo","amount":"1234"}`,
		//`{"denom":"bar","amount":"4321"}`,
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
	response := conn.out.String()
	fmt.Println(response)
	var output testpb.MsgRequest
	fmt.Println(output.U64)
	json.Unmarshal([]byte(response), &output)
	assert.Equal(t, output.GetU32(), uint32(27))
	//assert.Equal(t, output.GetU64(), uint64(5))
	assert.Equal(t, output.GetPositional1(), int32(1))
	assert.Equal(t, output.GetPositional2(), "abc")
}

func TestBuildCustomMsgCommand(t *testing.T) {
	b := &Builder{}
	customCommandCalled := false
	cmd, err := b.BuildMsgCommand(map[string]*autocliv1.ModuleOptions{
		"test": {
			Tx: testCmdMsgDesc,
		},
	}, map[string]*cobra.Command{
		"test": {Use: "test", Run: func(cmd *cobra.Command, args []string) {
			customCommandCalled = true
		}},
	})
	assert.NilError(t, err)
	cmd.SetArgs([]string{"test", "tx"})
	assert.NilError(t, cmd.Execute())
	assert.Assert(t, customCommandCalled)
}

func TestBuilder_BuildMsgMethodCommand(t *testing.T) {

}

type testMessageServer struct {
	testpb.UnimplementedMsgServer
}
